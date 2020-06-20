package learning

import (
	"github.com/Web-networks/execution-system/config"
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
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

func (spec *LearningTaskSpecification) GenerateWorkload(t *task.Task, _parameters task.Parameters) interface{} {
	parameters := _parameters.(task.LearningTaskParameters)
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
							Name:  "download-code",
							Image: "busybox",
							Command: []string{
								"wget",
								parameters.CodeUrl,
								"-O", "/neuroide/" + "code.tar.gz",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "resources",
									ReadOnly:  false,
									MountPath: "/neuroide",
								},
							},
						},
						{
							Name:  "untar-code",
							Image: "busybox",
							Command: []string{
								"tar",
								"-xzf", "/neuroide/" + "code.tar.gz",
								"-C", "/neuroide",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "resources",
									ReadOnly:  false,
									MountPath: "/neuroide",
								},
							},
						},
						{
							Name:  "learning",
							Image: "tensorflow/tensorflow",
							Command: []string{
								"python3", "/neuroide/cli.py",
								"--mode", "train",
								"--epochs", "5",
								"--sample-count", "100",
								"--weights", "/neuroide/weights.h5",
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
					Containers: []v1.Container{
						{
							Name:  "upload-weights",
							Image: "amazon/aws-cli",
							Command: []string{
								"aws",
								"--endpoint-url", "https://storage.yandexcloud.net",
								"--region", "ru-central1",
								"s3", "cp", "/neuroide/weights.h5", parameters.ResultS3Path,
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
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
}
