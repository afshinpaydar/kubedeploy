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

	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func getCurrentVersion(appName string) string {
	clientset, namespace := clientSet()
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), appName, metav1.GetOptions{})

	if err != nil {
		logger(err.Error(), Fatal)
	}

	version, ok := service.Spec.Selector["version"]
	if !ok {
		logger(fmt.Sprintf("FATAL: Unable to find current version deployed for %q", appName), Fatal)
	}
	return version
}

func switchOverService(appName string, version string) {
	clientset, namespace := clientSet()

	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		logger(err.Error(), Fatal)
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
		logger(err.Error(), Fatal)
	}
	logger(fmt.Sprintf("service/%s patched\n", service.Name), Info)
}

func findService(appName string) (name, sAppLabel, sVerLabel string) {
	clientset, namespace := clientSet()
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), appName, v1.GetOptions{})
	var sName, appLabel, verLabel string = "<Not Found>", "<Not Found>", "<Not Found>"

	if err == nil {
		sName = service.Name
	}

	if _, ok := service.Spec.Selector["app"]; ok {
		appLabel = service.Spec.Selector["app"]
	}

	if _, ok := service.Spec.Selector["version"]; ok {
		verLabel = service.Spec.Selector["version"]
	}
	return sName, appLabel, verLabel
}
