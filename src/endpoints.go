package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/Web-networks/execution-system/task"
	"github.com/gocraft/web"
)

type Context struct {
	Ctx context.Context
}

func NewEndpoints(manager *task.TaskManager) *Endpoints {
	return &Endpoints{
		taskManager: manager,
	}
}

type Endpoints struct {
	taskManager *task.TaskManager
}

func (ep *Endpoints) SetupRoutes(router *web.Router) {
	router.Middleware(web.LoggerMiddleware) // Use some included middleware
	router.Middleware(web.ShowErrorsMiddleware)
	router.Middleware(PanicRecoverMiddleware)

	router.Post("/api/task/:task_id/execute", ep.ExecuteTask)
	router.Get("/api/task/:task_id/state", ep.GetTaskState)
}

func (ep *Endpoints) ExecuteTask(ctx *Context, rw web.ResponseWriter, req *web.Request) {
	id, ok := req.PathParams["task_id"]
	if !ok {
		http.Error(rw, "task_id is not specified", http.StatusBadRequest)
	}

	var request ExecuteTaskRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, fmt.Sprintf("request is not a valid JSON: %v", err), http.StatusBadRequest)
	}

	t := task.NewTask(id, request.Type)

	err, data := ep.taskManager.Run(t)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	respBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	_, _ = rw.Write(respBytes)
}

type ExecuteTaskRequest struct {
	Type task.TaskType `json:"type"`
}

func (ep *Endpoints) GetTaskState(ctx *Context, rw web.ResponseWriter, req *web.Request) {
	id, ok := req.PathParams["task_id"]
	if !ok {
		http.Error(rw, "task_id is not specified", http.StatusBadRequest)
	}

	resp := GetTaskStatusResponse{
		State: ep.taskManager.TaskStateByID(id),
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	_, _ = rw.Write(respBytes)
}

type GetTaskStatusResponse struct {
	State task.TaskState `json:"state"`
}

func PanicRecoverMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			stack := make([]byte, size)
			stack = stack[:runtime.Stack(stack, false)]

			log.Printf("panic occured while serving request: %s\n%s", err, string(stack))
			http.Error(rw, fmt.Sprintf("Internal error: %s", err), http.StatusInternalServerError)
		}
	}()

	next(rw, req)
}
