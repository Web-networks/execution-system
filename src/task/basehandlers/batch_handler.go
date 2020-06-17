package basehandlers

import (
	"fmt"
	"log"

	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	batch "k8s.io/api/batch/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type BatchWorkloadTaskTypeHandler struct {
	kubeClient kube.Client
	spec       TaskSpecification
}

var _ task.TaskTypeHandler = (*BatchWorkloadTaskTypeHandler)(nil)

func NewBatchHandler(kubeClient kube.Client, spec TaskSpecification) *BatchWorkloadTaskTypeHandler {
	return &BatchWorkloadTaskTypeHandler{
		kubeClient: kubeClient,
		spec:       spec,
	}
}

func (h *BatchWorkloadTaskTypeHandler) Type() task.TaskType {
	return h.spec.Type()
}

func (h *BatchWorkloadTaskTypeHandler) RestoreTasks() ([]*task.Task, error) {
	jobs, err := h.restoreJobsFromKube()
	if err != nil {
		return nil, err
	}
	return h.tasksFromJobs(jobs), nil
}

func (h *BatchWorkloadTaskTypeHandler) tasksFromJobs(jobs *batch.JobList) []*task.Task {
	var tasks []*task.Task
	for _, job := range jobs.Items {
		tasks = append(tasks, newTaskFromWorkload(&job))
	}
	return tasks
}

func (h *BatchWorkloadTaskTypeHandler) restoreJobsFromKube() (*batch.JobList, error) {
	return h.kubeClient.GetBatchJobs(meta.ListOptions{
		// TODO: add managed-by
		LabelSelector: fmt.Sprintf("%s=%s", task.TaskTypeLabel, h.spec.Type()),
	})
}

func newTaskFromWorkload(job *batch.Job) *task.Task {
	t := &task.Task{
		ID:   idFromKubeJob(job),
		Type: typeFromKubeJob(job),
	}
	t.SetState(stateFromKubeJob(job))
	return t
}

func idFromKubeJob(job *batch.Job) string {
	labels := job.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[task.TaskIDLabel]
}

func typeFromKubeJob(job *batch.Job) task.TaskType {
	labels := job.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[task.TaskTypeLabel]
}

func stateFromKubeJob(job *batch.Job) task.TaskState {
	if job.Status.Failed > 0 {
		return task.Failed
	} else if job.Status.Active > 0 {
		return task.Running
	} else if job.Status.Succeeded > 0 {
		return task.Success
	} else {
		return task.Initializing
	}
}

func (h *BatchWorkloadTaskTypeHandler) WatchTasks(cb task.OnTaskStateModifiedCallback) {
	tasksWatcher, err := h.kubeClient.WatchBatchJobs(meta.ListOptions{
		// TODO: add managed-by
		LabelSelector: fmt.Sprintf("%s=%s", task.TaskTypeLabel, h.spec.Type()),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to start watching for %s tasks", h.spec.Type()))
	}

	go func() {
		log.Printf("watcher: start watching!")
		for event := range tasksWatcher.ResultChan() {
			switch event.Type {
			case watch.Modified:
				job := event.Object.DeepCopyObject().(*batch.Job)

				taskID := idFromKubeJob(job)
				newState := stateFromKubeJob(job)
				cb(taskID, newState)
			}
			// TODO: delete this
			log.Printf("watcher: event type = %v", event.Type)
		}
	}()
}

func (h *BatchWorkloadTaskTypeHandler) Run(task *task.Task) (error, string) {
	workload := h.spec.GenerateWorkload(task).(*batch.Job)
	err := h.kubeClient.RunBatchJob(workload)
	if err != nil {
		return err, ""
	}
	return nil, ""
}
