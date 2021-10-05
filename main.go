package main

import (
	"log"
	"net/http"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gomodule/redigo/redis"
	"github.com/sinmetal/gcpmetadata"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"google.golang.org/appengine/v2"
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

	http.Handle("/task/add", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(addTaskHandler) }(), "/gp"))
	http.Handle("/task/process", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(processTaskHandler) }(), "/gp"))
	http.Handle("/admin/hello", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(adminHandler) }(), "/gp"))
	http.Handle("/", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(CounterHandler) }(), "/gp"))

	appengine.Main()
}
