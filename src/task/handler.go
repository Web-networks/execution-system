package task

type TaskTypeHandler interface {
	Type() TaskType

	// startup
	RestoreTasks() ([]*Task, error)

	// runtime
	WatchTasks(cb OnTaskStateModifiedCallback)
	Run(task *Task) (error, string)
}

type OnTaskStateModifiedCallback = func(id string, newState TaskState)
