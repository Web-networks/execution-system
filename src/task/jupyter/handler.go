package jupyter

import (
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler(kubeClient kube.Client) task.TaskTypeHandler {
	return basehandlers.NewDeploymentHandler(kubeClient, &JupyterTaskSpecification{})
}

type JupyterTaskSpecification struct{}

var _ basehandlers.TaskWithServiceSpecification = (*JupyterTaskSpecification)(nil)

func (spec *JupyterTaskSpecification) Type() task.TaskType {
	return task.JupyterType
}

func (spec *JupyterTaskSpecification) GenerateWorkload(t *task.Task) interface{} {
	labels := map[string]string{
		task.ManagedByLabel: task.ManagedByValue,
		task.TaskTypeLabel:  task.JupyterType,
		task.TaskIDLabel:    t.ID,
	}
	return &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:   t.KubeWorkloadName(),
			Labels: labels,
		},
		Spec: apps.DeploymentSpec{
			Selector: &meta.LabelSelector{
				MatchLabels: labels,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: labels,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  t.Type,
							Image: "networksidea/image_generation:jupyter-1.0.5",
							Ports: []core.ContainerPort{
								{
									Name:          "http",
									Protocol:      core.ProtocolTCP,
									ContainerPort: 8888,
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "JUPYTER_TOKEN",
									Value: "abcd",
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyAlways,
					ImagePullSecrets: []core.LocalObjectReference{
						{
							Name: "dockerpullsecrets",
						},
					},
				},
			},
		},
	}
}

func (spec *JupyterTaskSpecification) GenerateService(t *task.Task) interface{} {
	labels := map[string]string{
		task.ManagedByLabel: task.ManagedByValue,
		task.TaskTypeLabel:  task.JupyterType,
		task.TaskIDLabel:    t.ID,
	}
	return &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:   t.KubeWorkloadName(),
			Labels: labels,
		},
		Spec: core.ServiceSpec{
			Selector: labels,
			Type:     "NodePort",
			Ports: []core.ServicePort{
				{
					Port: 8888,
				},
			},
		},
	}
}
