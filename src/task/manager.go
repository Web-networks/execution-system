package task

import (
	"log"

	"github.com/Web-networks/execution-system/kube"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type TaskManager struct {
	kubeClient kube.Client
	tasks      map[string]*Task // TaskID -> Task
}

func (tm *TaskManager) watchLearningTasks(event <-chan watch.Event) {
	log.Printf("watcher: start watching!")
	for event := range event {
		switch event.Type {
		case watch.Modified:
			// TODO: error handling
			job := event.Object.DeepCopyObject().(*v1.Job)
			taskID := idFromKubeJob(job)
			newState := stateFromKubeJob(job)
			tm.tasks[taskID].SetState(newState)
		}
		// TODO: MODIFIED - restarts
		log.Printf("watcher: event type = %v", event.Type)
	}
}

func (tm *TaskManager) Run(task *Task) error {
	err := tm.kubeClient.RunBatchJob(task.KubeJob())
	if err != nil {
		return err
	}
	tm.tasks[task.id] = task

	return nil
}

func (tm *TaskManager) TaskStateByID(id string) TaskState {
	foundTask, ok := tm.tasks[id]
	if !ok {
		return UnknownTask
	}
	return foundTask.State()
}
