package kube

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	clientset *kubernetes.Clientset
}

func NewClient(kubeConfigPath string) *client {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(errors.New(fmt.Sprintf("NewClient: failed to create config: %v", err)))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(errors.New(fmt.Sprintf("NewClient: failed to create clientset: %v", err)))
	}

	return &client{
		clientset: clientset,
	}
}
