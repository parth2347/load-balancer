package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	load_balancer "load-balancer"
	"net/http"
)

func main()  {
	r := mux.NewRouter()
	r.Use(load_balancer.WithLogging)
	proxy := load_balancer.Proxy{
		Host:    "localhost",
		Port:    3000,
		Scheme:  "http",
		Servers: []load_balancer.Server{{
			Host:        "localhost",
			Port:        "3000",
			Name:        "Server A",
			Scheme:      "http",
			Connections: 0,
		},
	}}
	r.HandleFunc("/", proxy.Handler)

	err := http.ListenAndServe(":8080", r)
	if err != nil{
		log.Error("Failed to bind the port with the process")
	}
}
