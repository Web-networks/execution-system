package basehandlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (h *BatchWorkloadTaskTypeHandler) tasksFromJobs(jobs *batchv1.JobList) []*task.Task {
	var tasks []*task.Task
	for _, job := range jobs.Items {
		tasks = append(tasks, newTaskFromWorkload(&job))
	}
	return tasks
}

func (h *BatchWorkloadTaskTypeHandler) restoreJobsFromKube() (*batchv1.JobList, error) {
	return h.kubeClient.GetBatchJobs(v1.ListOptions{
		// TODO: add managed-by
		LabelSelector: fmt.Sprintf("%s=%s", task.TaskTypeLabel, h.spec.Type()),
	})
}

func newTaskFromWorkload(job *batchv1.Job) *task.Task {
	t := &task.Task{
		ID:   idFromKubeJob(job),
		Type: typeFromKubeJob(job),
	}
	t.SetState(stateFromKubeJob(job))
	return t
}

func idFromKubeJob(job *batchv1.Job) string {
	labels := job.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[task.TaskIDLabel]
}

func typeFromKubeJob(job *batchv1.Job) task.TaskType {
	labels := job.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[task.TaskTypeLabel]
}

func stateFromKubeJob(job *batchv1.Job) task.TaskState {
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
	tasksWatcher, err := h.kubeClient.WatchBatchJobs(v1.ListOptions{
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
				job := event.Object.DeepCopyObject().(*batchv1.Job)

				if jobStatusString, err := json.Marshal(job.Status); err == nil {
					log.Printf("watcher: job.Status: %s", jobStatusString)
				} else {
					log.Printf("watcher: failed to marshal job.Status: %v", err)
				}

				taskID := idFromKubeJob(job)
				newState := stateFromKubeJob(job)
				cb(taskID, newState)
			}
			// TODO: delete this
			log.Printf("watcher: event type = %v", event.Type)
		}
	}()
}

func (h *BatchWorkloadTaskTypeHandler) Run(task *task.Task, parameters task.Parameters) error {
	workload := h.spec.GenerateWorkload(task, parameters).(*batchv1.Job)
	err := h.kubeClient.RunBatchJob(workload)
	if err != nil {
		return err
	}
	return nil
}
