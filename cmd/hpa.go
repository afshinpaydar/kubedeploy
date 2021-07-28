package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

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

func switchOverHPA(appName string, newDeploymentName string) {
	clientset, namespace := clientSet()
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/scaleTargetRef/name",
		Value: newDeploymentName,
	}}
	payloadBytes, _ := json.Marshal(payload)
	// Patch target of HPA
	_, err := clientset.AutoscalingV1().
		HorizontalPodAutoscalers(namespace).
		Patch(context.TODO(), appName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("horizontalpodautoscaler.autoscaling/%s patched\n", appName)
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

func calculateTargetReplica(appName string) int32 {
	var targetReplicas int32
	desiredReplicas := getDesiredReplicas(appName)
	minReplicas := getMinReplicas(appName)
	if desiredReplicas > minReplicas {
		targetReplicas = desiredReplicas
	} else {
		targetReplicas = minReplicas
	}
	return targetReplicas
}