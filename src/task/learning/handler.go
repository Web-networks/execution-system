package learning

import (
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler(kubeClient kube.Client) task.TaskTypeHandler {
	return basehandlers.NewBatchHandler(kubeClient, &LearningTaskSpecification{})
}

type LearningTaskSpecification struct{}

var _ basehandlers.TaskSpecification = (*LearningTaskSpecification)(nil)

func (spec *LearningTaskSpecification) Type() task.TaskType {
	return task.LearningType
}

func (spec *LearningTaskSpecification) GenerateWorkload(t *task.Task) interface{} {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.KubeJobName(),
			Labels: map[string]string{
				task.ManagedByLabel: task.ManagedByValue,
				task.TaskTypeLabel:  task.LearningType,
				task.TaskIDLabel:    t.ID,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "resources",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
					InitContainers: []v1.Container{
						{
							Name:    "resources",
							Image:   "busybox",
							Command: []string{"mkdir", "/neuroide/test"}, // create /neuroide dir
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "resources",
									ReadOnly:  false,
									MountPath: "/neuroide",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:    "learning",
							Image:   "busybox",
							Command: []string{"sleep", "100"}, // sleep for 30 seconds
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "resources",
									ReadOnly:  false,
									MountPath: "/neuroide",
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
