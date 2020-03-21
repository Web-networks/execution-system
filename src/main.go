package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Web-networks/execution-system/kube"
	"github.com/gocraft/web"
)

func main() {
	conf := NewConfig()

	kubeManager := kube.NewKubeClient(conf.KubeConfigPath)
	_ = kubeManager

	router := web.New(Context{}). // Create your router
					Middleware(web.LoggerMiddleware).    // Use some included middleware
					Middleware(web.ShowErrorsMiddleware) // ...

	log.Printf("Start listening on %s", conf.AddressAndPort())
	err := http.ListenAndServe(conf.AddressAndPort(), router)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Can't start server on %s: %v", conf.AddressAndPort(), err)))
	}
}
