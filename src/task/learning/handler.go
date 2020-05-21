package learning

import (
	"github.com/Web-networks/execution-system/config"
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const resourceDownloaderImageName = "asverdlov/resource-downloader"

func NewHandler(kubeClient kube.Client, config *config.Config) task.TaskTypeHandler {
	return basehandlers.NewBatchHandler(kubeClient, &LearningTaskSpecification{config: config})
}

type LearningTaskSpecification struct {
	config *config.Config
}

var _ basehandlers.TaskSpecification = (*LearningTaskSpecification)(nil)

func (spec *LearningTaskSpecification) Type() task.TaskType {
	return task.LearningType
}

func (spec *LearningTaskSpecification) GenerateWorkload(t *task.Task) interface{} {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.KubeWorkloadName(),
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
							Name:  "resource-downloader",
							Image: resourceDownloaderImageName,
							Command: []string{
								"/resource-downloader",
								"--s3_region", "ru-central1",
								"--s3_endpoint", "storage.yandexcloud.net",
								"--output_dir", "/neuroide",
								"--model_bucket", "code-testing",
								"--model_path", "model_a3ae012f-e1eb-4d0c-af36-b463e28cbaa1",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "resources",
									ReadOnly:  false,
									MountPath: "/neuroide",
								},
							},
							Env: []v1.EnvVar{
								{Name: "AWS_ACCESS_KEY_ID", Value: spec.config.AwsAccessKeyId},
								{Name: "AWS_SECRET_ACCESS_KEY", Value: spec.config.AwsSecretKey},
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
