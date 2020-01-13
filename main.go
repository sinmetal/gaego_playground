package main

import (
	"fmt"
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

func incrementHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, span := startSpan(ctx, "redis/increment")
	defer span.End()

	conn := redisPool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("warning: conn.Close() err=%+v\n", err)
		}
	}()

	counter, err := redis.Int(conn.Do("INCR", "visits"))
	if err != nil {
		http.Error(w, "Error incrementing visitor counter", http.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprintf(w, "Visitor number: %d", counter)
	if err != nil {
		fmt.Printf("failed write to response. err=%v\n", err)
	}
}

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

	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	const maxConnections = 10
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", redisAddr)
	}, maxConnections)

	http.Handle("/increment", ochttp.WithRouteTag(func() http.Handler { return http.HandlerFunc(incrementHandler) }(), "/gp"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
