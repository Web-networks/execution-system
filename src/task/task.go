package task

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

const ManagedByLabel = "app.kubernetes.io/managed-by"
const ManagedByValue = "execution-system"

const TaskTypeLabel = "bigone.demist.ru/task-type"
const TaskIDLabel = "bigone.demist.ru/task-id"

type TaskType = string

const (
	LearningType = "learning"
	ApplyingType = "applying"
	JupyterType  = "jupyter"
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
	ID   string
	Type TaskType

	state unsafe.Pointer
}

func NewTask(id string, typ TaskType) *Task {
	t := &Task{
		ID:   id,
		Type: typ,
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

func (t *Task) KubeJobName() string {
	return fmt.Sprintf("%s-%s", t.Type, t.ID)
}

func (t *Task) KubeDeploymentName() string {
	return fmt.Sprintf("%s-%s", t.Type, t.ID)
}
