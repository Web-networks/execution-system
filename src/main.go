package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Web-networks/execution-system/kube"
	"github.com/Web-networks/execution-system/task"
	"github.com/Web-networks/execution-system/task/learning"
	"github.com/gocraft/web"
)

func main() {
	conf := NewConfig()

	kubeClient := kube.NewClient(conf.KubeConfigPath)
	taskManager := task.CreateManagerFromKubernetesState(
		learning.NewTaskTypeHandler(kubeClient),
	)

	router := web.New(Context{})
	ep := NewEndpoints(taskManager)
	ep.SetupRoutes(router)

	log.Printf("Start listening on %s", conf.AddressAndPort())
	err := http.ListenAndServe(conf.AddressAndPort(), router)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Can't start server on %s: %v", conf.AddressAndPort(), err)))
	}
}
