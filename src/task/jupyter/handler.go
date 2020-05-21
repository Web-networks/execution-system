package jupyter

import (
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler(kubeClient kube.Client) task.TaskTypeHandler {
	return basehandlers.NewDeploymentHandler(kubeClient, &JupyterTaskSpecification{})
}

type JupyterTaskSpecification struct{}

var _ basehandlers.TaskSpecification = (*JupyterTaskSpecification)(nil)

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
		ObjectMeta: metav1.ObjectMeta{
			Name:   t.KubeWorkloadName(),
			Labels: labels,
		},
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    t.Type,
							Image:   "busybox",
							Command: []string{"sleep", "30"}, // sleep for 30 seconds
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}
}
