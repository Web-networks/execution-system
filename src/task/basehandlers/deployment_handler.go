package basehandlers

import (
	"fmt"
	"log"

	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type DeploymentHandler struct {
	kubeClient kube.Client
	spec       TaskSpecification
}

var _ task.TaskTypeHandler = (*DeploymentHandler)(nil)

func NewDeploymentHandler(kubeClient kube.Client, spec TaskSpecification) *DeploymentHandler {
	return &DeploymentHandler{
		kubeClient: kubeClient,
		spec:       spec,
	}
}

func (h *DeploymentHandler) Type() task.TaskType {
	return h.spec.Type()
}

func (h *DeploymentHandler) RestoreTasks() ([]*task.Task, error) {
	deployments, err := h.restoreDeploymentsFromKube()
	if err != nil {
		return nil, err
	}
	return h.tasksFromDeployments(deployments), nil
}

func (h *DeploymentHandler) tasksFromDeployments(deployments *apps.DeploymentList) []*task.Task {
	var tasks []*task.Task
	for _, deployment := range deployments.Items {
		t := &task.Task{
			ID:   idFromKubeDeployment(&deployment),
			Type: typeFromKubeDeployment(&deployment),
		}
		t.SetState(stateFromKubeDeployment(&deployment))
		tasks = append(tasks, t)
	}
	return tasks
}

func (h *DeploymentHandler) restoreDeploymentsFromKube() (*apps.DeploymentList, error) {
	return h.kubeClient.GetDeployments(v1.ListOptions{
		// TODO: add managed-by
		LabelSelector: fmt.Sprintf("%s=%s", task.TaskTypeLabel, h.spec.Type()),
	})
}

func idFromKubeDeployment(deployment *apps.Deployment) string {
	labels := deployment.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[task.TaskIDLabel]
}

func typeFromKubeDeployment(deployment *apps.Deployment) task.TaskType {
	labels := deployment.ObjectMeta.GetObjectMeta().GetLabels()
	return labels[task.TaskTypeLabel]
}

func stateFromKubeDeployment(deployment *apps.Deployment) task.TaskState {
	if deployment.Status.AvailableReplicas == deployment.Status.Replicas {
		return task.Running
	} else {
		return task.Initializing
	}
}

func (h *DeploymentHandler) WatchTasks(cb task.OnTaskStateModifiedCallback) {
	tasksWatcher, err := h.kubeClient.WatchDeployments(v1.ListOptions{
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
				deployment := event.Object.DeepCopyObject().(*apps.Deployment)

				taskID := idFromKubeDeployment(deployment)
				newState := stateFromKubeDeployment(deployment)
				cb(taskID, newState)
			}
		}
	}()
}

func (h *DeploymentHandler) Run(task *task.Task) error {
	workload := h.spec.GenerateWorkload(task).(*apps.Deployment)
	err := h.kubeClient.RunDeployment(workload)
	if err != nil {
		return err
	}
	return nil
}
