package applying

import (
	"fmt"

	"github.com/Web-networks/execution-system/config"
	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/basehandlers"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func NewHandler(kubeClient kube.Client, config *config.Config) task.TaskTypeHandler {
	return basehandlers.NewBatchHandler(kubeClient, &ApplyingTaskSpecification{config: config})
}

type ApplyingTaskSpecification struct {
	config *config.Config
}

var _ basehandlers.TaskSpecification = (*ApplyingTaskSpecification)(nil)

func (spec *ApplyingTaskSpecification) Type() task.TaskType {
	return task.ApplyingType
}

func (spec *ApplyingTaskSpecification) GenerateWorkload(t *task.Task, _parameters task.Parameters) interface{} {
	parameters := _parameters.(task.LearningOrApplyingTaskParameters)
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
			BackoffLimit: pointer.Int32Ptr(1),
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
							Name:  "download-code-and-weights",
							Image: "busybox",
							Command: []string{
								"/bin/sh",
								"-c",
								fmt.Sprintf(
									"wget %s -O /neuroide/code.tar.gz && wget %s -O /neuroide/weights.h5 && mkdir -p /neuroide/data/ && wget %s -O /neuroide/data/userdata",
									parameters.CodeUrl,
									parameters.WeightsUrl,
									parameters.DataUrl,
								),
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
							Name:  "applying",
							Image: "tensorflow/tensorflow",
							Command: []string{ // TODO
								"/bin/sh",
								"-c",
								"python3 /neuroide/cli.py --mode eval --eval-data /neuroide/data --weights /neuroide/weights.h5 --network-output /neuroide/results",
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
								"s3", "cp", "/neuroide/results", parameters.ResultS3Path,
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
