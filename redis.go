package main

import (
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
)

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
