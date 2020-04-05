package task

import (
	"fmt"
	"time"

	"github.com/Web-networks/execution-system/kube"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const restoreRetries = 10
const restoreTimeout = 1 * time.Second

func CreateManagerFromKubernetesState(kubeClient kube.Client) *TaskManager {
	jobList := restoreTasksFromKube(kubeClient)

	var tasks []*Task
	for _, job := range jobList.Items {
		tasks = append(tasks, newTaskFromBatchJob(&job))
	}

	return createManager(kubeClient, tasks)
}

func createManager(client kube.Client, tasks []*Task) *TaskManager {
	m := &TaskManager{
		kubeClient: client,
		tasks:      mapFromTasks(tasks),
	}

	learningTasksWatcher, err := m.kubeClient.WatchBatchJobs(v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", taskTypeLabel, LearningType),
	})
	if err != nil {
		panic("failed to start watching for learning tasks")
	}
	go m.watchLearningTasks(learningTasksWatcher.ResultChan())

	return m
}

func mapFromTasks(tasks []*Task) map[string]*Task {
	m := make(map[string]*Task)
	for _, task := range tasks {
		m[task.id] = task
	}
	return m
}

func tryRestoreTasksFromKube(kubeClient kube.Client) (*batchv1.JobList, error) {
	return kubeClient.GetBatchJobs(v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", managedByLabel, managedByValue),
	})
}

func restoreTasksFromKube(kubeClient kube.Client) *batchv1.JobList {
	for retry := 1; retry < restoreRetries; retry++ {
		jobList, err := tryRestoreTasksFromKube(kubeClient)
		if err == nil {
			return jobList
		}
		time.Sleep(restoreTimeout)
	}
	panic(fmt.Sprintf("failed to resync tasks with %d retries", restoreRetries))
}
