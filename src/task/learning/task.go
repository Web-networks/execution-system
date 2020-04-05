package learning

import (
	"github.com/Web-networks/execution-system/task"
	batchv1 "k8s.io/api/batch/v1"
)

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
