package task

import (
	"errors"
	"fmt"
)

type TaskManager struct {
	typeHandlers map[TaskType]TaskTypeHandler
	tasks        map[string]*Task // TaskID -> Task
}

func newTaskManager(tasks []*Task, handlers ...TaskTypeHandler) *TaskManager {
	m := &TaskManager{
		typeHandlers: mapFromTaskTypeHandlers(handlers),
		tasks:        mapFromTasks(tasks),
	}

	for _, handler := range handlers {
		handler.WatchTasks(m.onTaskStateChanged)
	}

	return m
}

func (tm *TaskManager) onTaskStateChanged(id string, state TaskState) {
	tm.tasks[id].SetState(state)
}

func (tm *TaskManager) Run(task *Task) error {
	handler, ok := tm.typeHandlers[task.Type]
	if !ok {
		return errors.New(fmt.Sprintf("unsupported task type '%s'", task.Type))
	}

	if _, ok := tm.tasks[task.ID]; ok {
		return errors.New(fmt.Sprintf("task with id %s already exists", task.ID))
	}

	err := handler.Run(task)
	if err != nil {
		return err
	}

	tm.tasks[task.ID] = task
	return nil
}

func (tm *TaskManager) TaskStateByID(id string) TaskState {
	foundTask, ok := tm.tasks[id]
	if !ok {
		return UnknownTask
	}
	return foundTask.State()
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
