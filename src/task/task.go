package task

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const K8sTypeLabel = "bigone.demist.ru/task-type"
const K8sTaskIDLabel = "bigone.demist.ru/task-id"

type TaskType = string

const (
	LearningType      = "learning"
	ApplyingType      = "applying"
	RemoteJupyterType = "jupyter"
)

type TaskState = string

const (
	Initializing = "INITIALIZING"
	Running      = "RUNNING"
	Failed       = "FAILED"
	Success      = "SUCCESS"
	UnknownTask  = "UNKNOWN_TASK"
)

type Task struct {
	id  string
	typ TaskType

	// TODO: mutex
	state TaskState
}

func NewTask(id string) *Task {
	return &Task{
		id:    id,
		typ:   LearningType,
		state: Initializing,
	}
}

func (t *Task) KubeJobName() string {
	return fmt.Sprintf("%s-%s", t.typ, t.id)
}

func (t *Task) KubeJob() *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.KubeJobName(),
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						K8sTypeLabel:   LearningType,
						K8sTaskIDLabel: t.id,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
		},
	}
}
