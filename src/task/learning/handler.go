package learning

import (
	"fmt"
	"log"

	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type handler struct {
	kubeClient kube.Client
}

var _ task.TaskTypeHandler = (*handler)(nil)

func NewTaskTypeHandler(kubeClient kube.Client) *handler {
	return &handler{
		kubeClient: kubeClient,
	}
}

func (h *handler) Type() task.TaskType {
	return task.LearningType
}

func (h *handler) RestoreTasks() ([]*task.Task, error) {
	jobs, err := h.restoreJobsFromKube()
	if err != nil {
		return nil, err
	}
	return h.tasksFromJobs(jobs), nil
}

func (h *handler) tasksFromJobs(jobs *batchv1.JobList) []*task.Task {
	var tasks []*task.Task
	for _, job := range jobs.Items {
		tasks = append(tasks, newTaskFromWorkload(&job))
	}
	return tasks
}

func (h *handler) restoreJobsFromKube() (*batchv1.JobList, error) {
	return h.kubeClient.GetBatchJobs(v1.ListOptions{
		// TODO: add managed-by
		LabelSelector: fmt.Sprintf("%s=%s", task.TaskTypeLabel, task.LearningType),
	})
}

func (h *handler) WatchTasks(cb task.OnTaskStateModifiedCallback) {
	learningTasksWatcher, err := h.kubeClient.WatchBatchJobs(v1.ListOptions{
		// TODO: add managed-by
		LabelSelector: fmt.Sprintf("%s=%s", task.TaskTypeLabel, task.LearningType),
	})
	if err != nil {
		panic("failed to start watching for learning tasks")
	}

	go func() {
		log.Printf("watcher: start watching!")
		for event := range learningTasksWatcher.ResultChan() {
			switch event.Type {
			case watch.Modified:
				job := event.Object.DeepCopyObject().(*batchv1.Job)

				taskID := idFromKubeJob(job)
				newState := stateFromKubeJob(job)
				cb(taskID, newState)
			}
			// TODO: delete this
			log.Printf("watcher: event type = %v", event.Type)
		}
	}()
}

func (h *handler) Run(task *task.Task) error {
	workload := generateWorkload(task)
	err := h.kubeClient.RunBatchJob(workload)
	if err != nil {
		return err
	}
	return nil
}
