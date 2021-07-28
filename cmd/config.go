package cmd

import (
	"fmt"
	"os"

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

func getTimeOut() int {
	timeout := 60
	if viper.Get("TimeOut") == nil {
		fmt.Println("Warning: TimeOut isn't set in config file: '$HOME/.kubectl-deploy.yaml', default is 60 Sec ")
	}

	timeout, ok := viper.Get("TimeOut").(int)
	if !ok {
		fmt.Printf("FATAL: Unexpected type for TimeOut Env: %v\n", viper.Get("TimeOut"))
		os.Exit(1)
	}
	return timeout
}

func getDockerHub() string {
	if viper.Get("DockerHub") == nil {
		fmt.Println("FATAL: DockerHub isn't set in config file: '$HOME/.kubectl-deploy.yaml' ")
		os.Exit(1)
	}
	dockerHub, ok := viper.Get("DockerHub").(string)
	if !ok {
		fmt.Printf("FATAL: Unexpected type for DockerHub Env: %v\n", viper.Get("DockerHub"))
		os.Exit(1)
	}
	return dockerHub
}

func getImageName() string {
	if viper.Get("ImageName") == nil {
		fmt.Println("FATAL: ImageName isn't set in config file: '$HOME/.kubectl-deploy.yaml' ")
		os.Exit(1)
	}
	imageName, ok := viper.Get("ImageName").(string)
	if !ok {
		fmt.Printf("FATAL: Unexpected type for ImageName Env: %v\n", viper.Get("ImageName"))
		os.Exit(1)
	}
	return imageName
}
