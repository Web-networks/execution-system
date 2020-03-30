package kube

import (
	"errors"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	RunBatchJob(job *batchv1.Job) error
	WatchBatchJobs(options v1.ListOptions) (watch.Interface, error)
}

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

func (c *client) RunBatchJob(job *batchv1.Job) error {
	jobClient := c.clientset.BatchV1().Jobs("default")
	_, err := jobClient.Create(job)
	return err
}

func (c *client) WatchBatchJobs(options v1.ListOptions) (watch.Interface, error) {
	jobClient := c.clientset.BatchV1().Jobs("default")
	return jobClient.Watch(options)
}
