/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const dormantVersion string = "dormant"

// bluegreenCmd represents the bluegreen command
var bluegreenCmd = &cobra.Command{
	Use:   "bluegreen",
	Short: "Simple blue/green deployment plugin for KUBECTL",
	Long: `"kube-deploy bluegreen" helps you to implement blue/green deployment in your k8s cluster
"kubectl-deploy bluegreen" expect two Deployments and one Service, that points to one of those in the active k8s cluster

the name of Deployments and Service doesn’t matter and could be anything,
and also how the Service exposed to outside of Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("FATAL: Please pass APP_NAME and VERSION as arguments")
			os.Exit(1)
		} else {
			blueGreenDeploy(args[0], args[1])
		}
	},
}

func init() {
	rootCmd.AddCommand(bluegreenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bluegreenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bluegreenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func clientSet() (*kubernetes.Clientset, string) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return clientset, namespace
}

func getCurrentVersion(appName string) string {
	clientset, namespace := clientSet()
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), appName, metav1.GetOptions{})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	version, ok := service.Spec.Selector["version"]
	if !ok {
		fmt.Printf("FATAL: Unable to find current version deployed for '%s'", appName)
		os.Exit(1)
	}
	return version
}

func getOldDeployment(appName string) string {

	clientset, namespace := clientSet()
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": getCurrentVersion(appName), "app": appName}}
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:         1,
	}
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("FATAL: Unable to find live deployment for '%s' version\n", getCurrentVersion(appName))
			os.Exit(1)
		}
	}()

	oldDeployment := deployments.Items[0].Name
	return oldDeployment
}

func getNewDeployment(appName string) string {

	clientset, namespace := clientSet()
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": dormantVersion, "app": appName}}
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:         1,
	}
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("FATAL: Unable to find spare dormant deploymen")
			os.Exit(1)
		}
	}()

	newDeployment := deployments.Items[0].Name
	return newDeployment
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func getDesiredReplicas(appName string) int32 {
	clientset, namespace := clientSet()

	hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Get(context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Warning: Unable to find disired replicas for Horizontal Pod Autoscaler")
			os.Exit(1)
		}
	}()

	disiredReplicas := hpa.Status.DesiredReplicas
	return int32(disiredReplicas)
}

func getMinReplicas(appName string) int32 {
	clientset, namespace := clientSet()

	hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Get(context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Warning: Unable to find min replicas for Horizontal Pod Autoscaler")
			os.Exit(1)
		}
	}()

	minReplicas := hpa.Spec.MinReplicas
	return int32(*minReplicas)
}

func blueGreenDeploy(appName string, version string) {
	newDeploymentName := getNewDeployment(appName)
	oldDeploymentName := getOldDeployment(appName)
	if newDeploymentName == oldDeploymentName {
		fmt.Println("FATAL: Something is wrong, Dormant and Live are somehow the same")
		os.Exit(1)
	}
	clientset, namespace := clientSet()
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/template/metadata/labels/version",
		Value: version,
	}}
	payloadBytes, _ := json.Marshal(payload)
	// Patch label of Deployment
	_, err := clientset.
		AppsV1().Deployments(namespace).
		Patch(context.TODO(), newDeploymentName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if viper.Get("DockerOrg") == nil {
		fmt.Println("FATAL: DockerOrg isn't set in config file: '$HOME/.kubectl-deploy.yaml' ")
		os.Exit(1)
	}

	if viper.Get("ImageName") == nil {
		fmt.Println("FATAL: ImageName isn't set in config file: '$HOME/.kubectl-deploy.yaml' ")
		os.Exit(1)
	}
	dockerHubOrg, ok := viper.Get("DockerOrg").(string)
	if !ok {
		fmt.Printf("FATAL: Unexpected type for DockerOrg Env: %v\n", viper.Get("DockerOrg"))
		os.Exit(1)
	}

	imageName, ok := viper.Get("ImageName").(string)
	if !ok {
		fmt.Printf("FATAL: Unexpected type for ImageName Env: %v\n", viper.Get("DockerOrg"))
		os.Exit(1)
	}

	payloadTemplate := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/template/spec/containers/0/image",
		Value: fmt.Sprintf("%s/%s:%s", dockerHubOrg, imageName, version),
	}}
	payloadBytesTemplate, _ := json.Marshal(payloadTemplate)

	// Patch image of Template
	_, errTemplate := clientset.
		AppsV1().Deployments(namespace).
		Patch(context.TODO(), newDeploymentName, types.JSONPatchType, payloadBytesTemplate, metav1.PatchOptions{})

	if errTemplate != nil {
		fmt.Println(errTemplate)
		os.Exit(1)
	} else {
		fmt.Printf("deployment.apps/%s patched\n", newDeploymentName)
	}

	payloadDeployment := []patchStringValue{{
		Op:    "replace",
		Path:  "/metadata/labels/version",
		Value: version,
	}}
	payloadBytesDeployment, _ := json.Marshal(payloadDeployment)

	// Patch image of Deployment
	_, errDeployment := clientset.
		AppsV1().Deployments(namespace).
		Patch(context.TODO(), newDeploymentName, types.JSONPatchType, payloadBytesDeployment, metav1.PatchOptions{})
	if errDeployment != nil {
		fmt.Println(errDeployment)
		os.Exit(1)
	}

	// Calculate targetReplica
	var targetReplicas int32
	desiredReplicas := getDesiredReplicas(appName)
	minReplicas := getMinReplicas(appName)
	if desiredReplicas > minReplicas {
		targetReplicas = desiredReplicas
	} else {
		targetReplicas = minReplicas
	}

	// Scale new deployment to targetReplicas
	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newDeploymentName,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: targetReplicas,
		},
	}
	_, err = clientset.
		AppsV1().Deployments(namespace).
		UpdateScale(context.TODO(), newDeploymentName, scale, metav1.UpdateOptions{})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Printf("deployment.apps/%s scaled\n", newDeploymentName)
	}

	rolloutStatus := waitRolloutStatus(newDeploymentName, appName, targetReplicas)
	if !rolloutStatus {
		fmt.Println("FATAL: Rollout of new version failed! Release aborted.")
		// TODO: deploymentStatus

		// Set the new deployment to be dormant again ready for the next release
		payloadTemplate := []patchStringValue{{
			Op:    "replace",
			Path:  "/spec/template/metadata/labels/version",
			Value: "dormant",
		}}
		payloadBytesTemplate, _ := json.Marshal(payloadTemplate)
		// Patch label of Template
		_, errTemplate := clientset.
			AppsV1().Deployments(namespace).
			Patch(context.TODO(), newDeploymentName, types.JSONPatchType, payloadBytesTemplate, metav1.PatchOptions{})
		if errTemplate != nil {
			fmt.Println(errTemplate)
			os.Exit(1)
		}

		payloadDeployment := []patchStringValue{{
			Op:    "replace",
			Path:  "/metadata/labels/version",
			Value: "dormant",
		}}
		payloadBytesDeployment, _ := json.Marshal(payloadDeployment)
		// Patch label of Deployment
		_, errDeployment := clientset.
			AppsV1().Deployments(namespace).
			Patch(context.TODO(), newDeploymentName, types.JSONPatchType, payloadBytesDeployment, metav1.PatchOptions{})
		if errDeployment != nil {
			fmt.Println(errDeployment)
			os.Exit(1)
		}

		os.Exit(1)
	}

}

func waitRolloutStatus(newDeploymentName string, appName string, targetReplicas int32) bool {
	clientset, namespace := clientSet()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("FATAL: Unable to find spare dormant deploymen")
			os.Exit(1)
		}
	}()

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": "dormant", "app": appName}}
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:         1,
	}

	timeout := 60
	var ok bool
	if viper.Get("TimeOut") == nil {
		fmt.Println("Warning: TimeOut isn't set in config file: '$HOME/.kubectl-deploy.yaml', default is 60 Sec ")
	} else {
		timeout, ok = viper.Get("TimeOut").(int)
		if !ok {
			fmt.Printf("FATAL: Unexpected type for TimeOut Env: %v\n", viper.Get("TimeOut"))
			os.Exit(1)
		}
	}

	for {
		deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if deployments.Items[0].Status.ReadyReplicas < targetReplicas && timeout > 0 {
			fmt.Printf("Waiting for deployment '%s' rollout to finish: %d out of %d new replicas have been updated...\n", newDeploymentName, deployments.Items[0].Status.ReadyReplicas, targetReplicas)
			timeout -= 5
			time.Sleep(5 * time.Second)
			continue
		} else if timeout <= 0 {
			fmt.Printf("deployment '%s' rollout timeout\n", newDeploymentName)
			return false
		} else if deployments.Items[0].Status.ReadyReplicas == targetReplicas {
			fmt.Printf("deployment '%s' successfully rolled out\n", newDeploymentName)
			return true
		}
	}
}
