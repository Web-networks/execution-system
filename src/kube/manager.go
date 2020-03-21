package kube

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	clientset *kubernetes.Clientset
}

func NewKubeClient(kubeConfigPath string) *KubeClient {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(errors.New(fmt.Sprintf("NewKubeClient: failed to create config: %v", err)))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(errors.New(fmt.Sprintf("NewKubeClient: failed to create clientset: %v", err)))
	}

	return &KubeClient{
		clientset: clientset,
	}
}
