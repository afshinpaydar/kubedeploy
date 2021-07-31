/*
Copyright © 2021 Afshin Paydar <afshinpaydar@gmail.com>

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

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
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
			logger(fmt.Sprintf("Unable to find live deployment for '%s' version\n", getCurrentVersion(appName)), Fatal)
		}
	}()

	oldDeployment := deployments.Items[0].Name
	return oldDeployment
}

func getNewDeployment(appName string) string {

	clientset, namespace := clientSet()
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": "dormant", "app": appName}}
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
			logger(fmt.Sprintf("Deployment name or tags are wrong, run 'kubectl deploy show %s' for more details", appName), Fatal)
		}
	}()

	return deployments.Items[0].Name
}

func blueGreenDeploy(appName string, version string) {
	newDeploymentName := getNewDeployment(appName)
	oldDeploymentName := getOldDeployment(appName)
	if newDeploymentName == oldDeploymentName {
		logger(fmt.Sprintf("Deployment name or tags are wrong, run 'kubectl deploy show %s' for more details", appName), Fatal)
	}

	if getCurrentVersion(appName) == version {
		logger(fmt.Sprintf("App '%s' with version '%s' already exists, run 'kubectl deploy show %s' for more details", appName, version, appName), Fatal)
	}

	dockerHub := getDockerHub()
	imageName := getImageName()
	// Patch new deployment to new version and label
	patchDeployment(newDeploymentName, version, dockerHub, imageName)

	targetReplicas := calculateTargetReplica(appName)
	// Scale up new deployment to targetReplicas
	scaleDeployment(newDeploymentName, targetReplicas)

	rolloutStatus := waitRolloutStatus(newDeploymentName, appName, targetReplicas, version)

	if !rolloutStatus {
		// TODO: deploymentStatus
		logger("Rollout of new version failed! Release aborted.", Fatal)
	}

	switchOverService(appName, version)
	switchOverHPA(appName, newDeploymentName)
	// Scale down old deployment to zero
	scaleDeployment(oldDeploymentName, 0)
	// Set the old deployment to be dormant ready for the next release
	patchDeployment(oldDeploymentName, "dormant", dockerHub, imageName)
	logger("Success: Release complete", Info)
}

func waitRolloutStatus(deploymentName string, appName string, targetReplicas int32, version string) bool {
	dockerHub := getDockerHub()
	imageName := getImageName()

	clientset, namespace := clientSet()
	defer func() {
		if r := recover(); r != nil {
			// TODO: Rollback dormant and scale down to zero
			scaleDeployment(deploymentName, 0)
			patchDeployment(deploymentName, "dormant", dockerHub, imageName)
			logger("Unable to find spare deploymen", Fatal)
		}
	}()

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": version, "app": appName}}
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:         1,
	}

	timeout := getTimeOut()
	for {
		deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if deployments.Items[0].Status.ReadyReplicas < targetReplicas && timeout > 0 {
			fmt.Printf("Waiting for deployment '%s' rollout to finish: %d out of %d new replicas have been updated...\n", deploymentName, deployments.Items[0].Status.ReadyReplicas, targetReplicas)
			timeout -= 5
			time.Sleep(5 * time.Second)
			continue
		} else if timeout <= 0 {
			// Set the new deployment to be dormant again ready for the next release
			scaleDeployment(deploymentName, 0)
			patchDeployment(deploymentName, "dormant", dockerHub, imageName)
			logger(fmt.Sprintf("deployment '%s' rollout timeout\n", deploymentName), Warn)
			return false
		} else if deployments.Items[0].Status.ReadyReplicas == targetReplicas {
			fmt.Printf("deployment '%s' successfully rolled out to version '%s'\n", deploymentName, version)
			return true
		} else {
			fmt.Printf("deployment '%s' rollout faild\n", deploymentName)
			return false
		}
	}
}

func scaleDeployment(deploymentName string, replica int32) {
	clientset, namespace := clientSet()

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replica,
		},
	}
	_, err := clientset.
		AppsV1().Deployments(namespace).
		UpdateScale(context.TODO(), deploymentName, scale, metav1.UpdateOptions{})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("deployment.apps/%s scaled\n", deploymentName)
}

func patchDeployment(deploymentName string, version string, dockerHub string, imageName string) {
	clientset, namespace := clientSet()

	// Patch lable of Template
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/template/metadata/labels/version",
		Value: version,
	}}
	payloadBytes, _ := json.Marshal(payload)
	// Patch label of Deployment
	_, err := clientset.
		AppsV1().Deployments(namespace).
		Patch(context.TODO(), deploymentName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	payloadTemplate := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/template/spec/containers/0/image",
		Value: fmt.Sprintf("%s/%s:%s", dockerHub, imageName, version),
	}}
	payloadBytesTemplate, _ := json.Marshal(payloadTemplate)

	// Patch image of Template
	_, errT := clientset.
		AppsV1().Deployments(namespace).
		Patch(context.TODO(), deploymentName, types.JSONPatchType, payloadBytesTemplate, metav1.PatchOptions{})

	if errT != nil {
		fmt.Println(errT)
		os.Exit(1)
	}

	payloadDeployment := []patchStringValue{{
		Op:    "replace",
		Path:  "/metadata/labels/version",
		Value: version,
	}}
	payloadBytesDeployment, _ := json.Marshal(payloadDeployment)

	// Patch labels of Deployment
	_, errD := clientset.
		AppsV1().Deployments(namespace).
		Patch(context.TODO(), deploymentName, types.JSONPatchType, payloadBytesDeployment, metav1.PatchOptions{})
	if errD != nil {
		fmt.Println(errD)
		os.Exit(1)
	}
	fmt.Printf("deployment.apps/%s patched\n", deploymentName)
}

func findDeployment(appName string, color string) (deployName, gDeployAppLabel, deployVerLabel string) {
	clientset, namespace := clientSet()

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), appName+"-"+color, v1.GetOptions{})
	if err != nil {
		deployName = "<Not Found>"
	} else {
		deployName = deployment.Name
	}

	deployAppLabel, ok := deployment.Labels["app"]
	if !ok {
		deployAppLabel = "<Not Found>"
	}
	deployVerLabel, ok = deployment.Labels["version"]
	if !ok {
		deployVerLabel = "<Not Found>"
	}
	return deployName, deployAppLabel, deployVerLabel
}
