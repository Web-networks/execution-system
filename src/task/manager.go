package task

import (
	"fmt"
	"log"

	"github.com/Web-networks/execution-system/kube"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func NewTaskManager(client kube.Client) *TaskManager {
	m := &TaskManager{
		kubeClient: client,
		tasks:      make(map[string]*Task),
	}

	remoteJupyterWatcher, err := m.kubeClient.WatchBatchJobs(v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", K8sTypeLabel, LearningType),
	})
	if err != nil {
		panic("failed to start watching for learning tasks")
	}
	go m.watchLearningTasks(remoteJupyterWatcher.ResultChan())

	return m
}

type TaskManager struct {
	kubeClient kube.Client
	tasks      map[string]*Task // TaskID -> Task
}

func (tm *TaskManager) watchLearningTasks(event <-chan watch.Event) {
	log.Printf("watcher: start watching!")
	for event := range event {
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
	return foundTask.state
}
