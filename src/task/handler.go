package task

type TaskTypeHandler interface {
	Type() TaskType

	// startup
	RestoreTasks() ([]*Task, error)

	// runtime
	WatchTasks(cb OnTaskStateModifiedCallback)
	Run(task *Task, parameters Parameters) error
}

type OnTaskStateModifiedCallback = func(id string, newState TaskState)
