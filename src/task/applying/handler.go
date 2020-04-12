package applying

import (
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler(kubeClient kube.Client) task.TaskTypeHandler {
	return basehandlers.NewBatchHandler(kubeClient, &ApplyingTaskSpecification{})
}

type ApplyingTaskSpecification struct{}

var _ basehandlers.TaskSpecification = (*ApplyingTaskSpecification)(nil)

func (spec *ApplyingTaskSpecification) Type() task.TaskType {
	return task.ApplyingType
}

func (spec *ApplyingTaskSpecification) GenerateWorkload(t *task.Task) interface{} {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.KubeJobName(),
			Labels: map[string]string{
				task.ManagedByLabel: task.ManagedByValue,
				task.TaskTypeLabel:  task.ApplyingType,
				task.TaskIDLabel:    t.ID,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "applying",
							Image:   "busybox",
							Command: []string{"sleep", "10"}, // sleep for 30 seconds
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
}
