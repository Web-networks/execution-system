package task

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const managedByLabel = "app.kubernetes.io/managed-by"
const managedByValue = "execution-system"

const taskTypeLabel = "bigone.demist.ru/task-type"
const taskIDLabel = "bigone.demist.ru/task-id"

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

	UnknownTask = "UNKNOWN_TASK"
)

type Task struct {
	id  string
	typ TaskType

	state unsafe.Pointer
}

func NewTask(id string) *Task {
	t := &Task{
		id:  id,
		typ: LearningType,
	}
	t.SetState(Initializing)
	return t
}

func (t *Task) State() TaskState {
	state := atomic.LoadPointer(&t.state)
	return *(*TaskState)(state)
}

func (t *Task) SetState(state TaskState) {
	atomic.StorePointer(&t.state, unsafe.Pointer(&state))
}

func newTaskFromBatchJob(job *batchv1.Job) *Task {
	t := &Task{
		id:  idFromKubeJob(job),
		typ: typeFromKubeJob(job),
	}
	t.SetState(stateFromKubeJob(job))
	return t
}

func (t *Task) KubeJobName() string {
	return fmt.Sprintf("%s-%s", t.typ, t.id)
}

func (t *Task) KubeJob() *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.KubeJobName(),
			Labels: map[string]string{
				managedByLabel: managedByValue,
				taskTypeLabel:  LearningType,
				taskIDLabel:    t.id,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
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

func idFromKubeJob(job *batchv1.Job) string {
	labels := job.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[taskIDLabel]
}

func typeFromKubeJob(job *batchv1.Job) TaskType {
	labels := job.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[taskTypeLabel]
}

func stateFromKubeJob(job *batchv1.Job) TaskState {
	if job.Status.Failed > 0 {
		return Failed
	} else if job.Status.Active > 0 {
		return Running
	} else if job.Status.Succeeded > 0 {
		return Success
	} else {
		return Initializing
	}
}
