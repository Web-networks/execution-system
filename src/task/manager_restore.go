package task

import (
	"fmt"
	"time"

	"github.com/Web-networks/execution-system/kube"
)

const restoreRetries = 10
const restoreTimeout = 1 * time.Second

func CreateManagerFromKubernetesState(kubeClient kube.Client, handlers ...TaskTypeHandler) *TaskManager {
	var restoredTasks []*Task

	for _, handler := range handlers {
		restoredTasks = append(restoredTasks, restoreTasks(handler)...)
	}

	return newTaskManager(kubeClient, restoredTasks, handlers...)
}

func restoreTasks(handler TaskTypeHandler) []*Task {
	for retry := 1; retry < restoreRetries; retry++ {
		tasks, err := handler.RestoreTasks()
		if err == nil {
			return tasks
		}
		time.Sleep(restoreTimeout)
	}
	panic(fmt.Sprintf("failed to resync tasks with %d retries", restoreRetries))
}
