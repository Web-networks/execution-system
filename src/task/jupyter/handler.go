package jupyter

import (
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler(kubeClient kube.Client) task.TaskTypeHandler {
	return basehandlers.NewBatchHandler(kubeClient, &JupyterTaskSpecification{})
}

type JupyterTaskSpecification struct{}

var _ basehandlers.TaskSpecification = (*JupyterTaskSpecification)(nil)

func (spec *JupyterTaskSpecification) Type() task.TaskType {
	return task.JupyterType
}

func (spec *JupyterTaskSpecification) GenerateWorkload(t *task.Task) interface{} {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.KubeJobName(),
			Labels: map[string]string{
				task.ManagedByLabel: task.ManagedByValue,
				task.TaskTypeLabel:  task.JupyterType,
				task.TaskIDLabel:    t.ID,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "jupyter",
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
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
}
