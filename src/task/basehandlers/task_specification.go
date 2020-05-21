package basehandlers

import "github.com/Web-networks/execution-system/task"

type TaskSpecification interface {
	Type() task.TaskType
	GenerateWorkload(t *task.Task) interface{}
}
type TaskWithServiceSpecification interface {
	Type() task.TaskType
	GenerateWorkload(t *task.Task) interface{}
	GenerateService(t *task.Task) interface{}
}
