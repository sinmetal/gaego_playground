package main

import (
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gomodule/redigo/redis"
	"github.com/sinmetal/gcpmetadata"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

var redisPool *redis.Pool

func main() {
	project, err := gcpmetadata.GetProjectID()
	if err != nil {
		log.Printf("ProjectID not found")
	}

	if project != "" {
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: project,
		})
		if err != nil {
			panic(err)
		}
		trace.RegisterExporter(exporter)
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	// redisが存在する時だけ動かす
	if setupRedis() {
		http.Handle("/increment", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(incrementHandler) }(), "/gp"))
	}

	http.Handle("/admin/hello", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(adminHandler) }(), "/gp"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
