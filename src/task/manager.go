package task

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const tasksFullSyncInterval = 10 * time.Second

type TaskManager struct {
	typeHandlers map[TaskType]TaskTypeHandler

	mu    sync.Mutex
	tasks map[string]*Task // TaskID -> Task
}

func newTaskManager(tasks []*Task, handlers ...TaskTypeHandler) *TaskManager {
	m := &TaskManager{
		typeHandlers: mapFromTaskTypeHandlers(handlers),
		tasks:        mapFromTasks(tasks),
	}

	for _, handler := range handlers {
		handler.WatchTasks(m.onTaskStateChanged)
	}
	go m.PeriodicallySyncTasks(time.NewTicker(tasksFullSyncInterval))

	return m
}

func (tm *TaskManager) onTaskStateChanged(id string, state TaskState) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tasks[id].SetState(state)
}

func (tm *TaskManager) Run(task *Task, parameters Parameters) error {
	handler, ok := tm.typeHandlers[task.Type]
	if !ok {
		return errors.New(fmt.Sprintf("unsupported task type '%s'", task.Type))
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, ok := tm.tasks[task.ID]; ok {
		return errors.New(fmt.Sprintf("task with id %s already exists", task.ID))
	}

	err := handler.Run(task, parameters)
	if err != nil {
		return err
	}

	tm.tasks[task.ID] = task
	return nil
}

func (tm *TaskManager) TaskStateByID(id string) TaskState {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	foundTask, ok := tm.tasks[id]
	if !ok {
		return UnknownTask
	}
	return foundTask.State()
}

func (tm *TaskManager) ListTasks() []*Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var tasks []*Task
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// sometimes change events are not passed into watchers
func (tm *TaskManager) PeriodicallySyncTasks(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			var restoredTasks []*Task
			for _, handler := range tm.typeHandlers {
				restoredTasks = append(restoredTasks, restoreTasks(handler)...)
			}

			tm.mu.Lock()
			tm.tasks = mapFromTasks(restoredTasks)
			tm.mu.Unlock()
		}
		log.Printf("Successfully synced tasks")
	}
}

func mapFromTasks(tasks []*Task) map[string]*Task {
	m := make(map[string]*Task)
	for _, task := range tasks {
		m[task.ID] = task
	}
	return m
}

func mapFromTaskTypeHandlers(handlers []TaskTypeHandler) map[TaskType]TaskTypeHandler {
	m := make(map[TaskType]TaskTypeHandler)
	for _, handler := range handlers {
		m[handler.Type()] = handler
	}
	return m
}
