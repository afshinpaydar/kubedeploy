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
	"time"

	"context"

	"github.com/briandowns/spinner"
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
		logger(err.Error(), Fatal)
	}

	defer func() {
		if r := recover(); r != nil {
			logger(fmt.Sprintf("Unable to find live deployment for %q version\n", getCurrentVersion(appName)), Fatal)
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
		logger(err.Error(), Fatal)
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
		logger(fmt.Sprintf("App %q with version %q already exists, run 'kubectl deploy show %s' for more details", appName, version, appName), Fatal)
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
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Color("green")
	for {
		deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), opts)
		s.Prefix = fmt.Sprintf("Waiting for deployment %q rollout ", deploymentName)
		s.Restart()
		if err != nil {
			logger(err.Error(), Fatal)
		}
		if deployments.Items[0].Status.ReadyReplicas < targetReplicas && timeout > 0 {
			timeout -= 2
			time.Sleep(2 * time.Second)
			continue
		}
		if timeout <= 0 {
			// Set the new deployment to be dormant again ready for the next release
			s.Stop()
			fmt.Println("")
			scaleDeployment(deploymentName, 0)
			patchDeployment(deploymentName, "dormant", dockerHub, imageName)
			logger(fmt.Sprintf("deployment %q rollout timeout\n", deploymentName), Warn)
			return false
		}
		if deployments.Items[0].Status.ReadyReplicas == targetReplicas {
			s.Stop()
			fmt.Println("")
			logger(fmt.Sprintf("deployment %q successfully rolled out to version %q\n", deploymentName, version), Info)
		}
		return true
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
		logger(err.Error(), Fatal)
	}

	logger(fmt.Sprintf("deployment.apps/%s scaled\n", deploymentName), Info)
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
		logger(err.Error(), Fatal)
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
		logger(errT.Error(), Fatal)
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
		logger(errD.Error(), Fatal)
	}

	logger(fmt.Sprintf("deployment.apps/%s patched\n", deploymentName), Info)
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
