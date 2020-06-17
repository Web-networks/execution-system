package kube

import (
	"errors"
	"fmt"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	RunBatchJob(job *batch.Job) error
	WatchBatchJobs(options meta.ListOptions) (watch.Interface, error)
	GetBatchJobs(options meta.ListOptions) (*batch.JobList, error)
	RunDeployment(deployment *apps.Deployment) error
	WatchDeployments(options meta.ListOptions) (watch.Interface, error)
	GetDeployments(options meta.ListOptions) (*apps.DeploymentList, error)
	CreateService(service *core.Service) (*core.Service, error)
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

func (c *client) RunBatchJob(job *batch.Job) error {
	jobClient := c.clientset.BatchV1().Jobs("default")
	_, err := jobClient.Create(job)
	return err
}

func (c *client) WatchBatchJobs(options meta.ListOptions) (watch.Interface, error) {
	jobClient := c.clientset.BatchV1().Jobs("default")
	return jobClient.Watch(options)
}

func (c *client) GetBatchJobs(options meta.ListOptions) (*batch.JobList, error) {
	jobClient := c.clientset.BatchV1().Jobs("default")
	return jobClient.List(options)
}

func (c *client) RunDeployment(deployment *apps.Deployment) error {
	deploymentClient := c.clientset.AppsV1().Deployments("default")
	_, err := deploymentClient.Create(deployment)
	return err
}

func (c *client) WatchDeployments(options meta.ListOptions) (watch.Interface, error) {
	deploymentClient := c.clientset.AppsV1().Deployments("default")
	return deploymentClient.Watch(options)
}

func (c *client) GetDeployments(options meta.ListOptions) (*apps.DeploymentList, error) {
	deploymentClient := c.clientset.AppsV1().Deployments("default")
	return deploymentClient.List(options)
}

func (c *client) CreateService(service *core.Service) (*core.Service, error) {
	serviceClient := c.clientset.CoreV1().Services("default")
	createdService, err := serviceClient.Create(service)
	//log.Print(createdService.Spec.Ports[0].NodePort)
	return createdService, err
}
