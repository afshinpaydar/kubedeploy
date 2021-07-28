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

func switchOverService(appName string, version string) {
	clientset, namespace := clientSet()

	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("FATAL: Unable to find service")
		os.Exit(1)
	}

	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/selector/version",
		Value: version,
	}}
	payloadBytes, _ := json.Marshal(payload)
	// Patch label of Service
	_, err = clientset.CoreV1().
		Services(namespace).
		Patch(context.TODO(), service.Name, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("service/%s patched\n", service.Name)
}
