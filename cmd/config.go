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
	"fmt"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func clientSet() (*kubernetes.Clientset, string) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		logger(err.Error(), Fatal)
	}
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger(err.Error(), Fatal)
	}
	return clientset, namespace
}

func getTimeOut() int {
	timeout := 60
	if viper.Get("TimeOut") == nil {
		fmt.Println("Warning: TimeOut isn't set in config file: '$HOME/.kubectl-deploy.yaml', default is 60 Sec ")
	}

	timeout, ok := viper.Get("TimeOut").(int)
	if !ok {
		logger(fmt.Sprintf("Unexpected type for TimeOut Env: %v\n", viper.Get("TimeOut")), Fatal)
	}
	return timeout
}

func getDockerHub() string {
	if viper.Get("DockerRegistryName") == nil {
		logger("'DockerRegistryName' isn't set in config file: '$HOME/.kubectl-deploy.yaml'", Fatal)
	}
	DockerRegistryName, ok := viper.Get("DockerRegistryName").(string)
	if !ok {
		logger(fmt.Sprintf("Unexpected type for DockerRegistryName Env: %v\n", viper.Get("DockerRegistryName")), Fatal)
	}
	return DockerRegistryName
}

func getImageName() string {
	if viper.Get("ImageName") == nil {
		logger("ImageName isn't set in config file: '$HOME/.kubectl-deploy.yaml' ", Fatal)
	}
	imageName, ok := viper.Get("ImageName").(string)
	if !ok {
		logger(fmt.Sprintf("Unexpected type for ImageName Env: %v\n", viper.Get("ImageName")), Fatal)
	}
	return imageName
}
