package main

import (
	"fmt"
	"net/http"
	"time"

	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/log"
	"google.golang.org/appengine/v2/memcache"
)

func CounterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	const key = "counter"
	var count int64

	v, err := memcache.Get(ctx, key)
	if err == memcache.ErrNotStored {
	} else if err != nil {
		log.Infof(ctx, "failed memcache get counter %s", err)
	}

	if v != nil {
		count = v.Object.(int64)
	}

	err = memcache.Add(ctx, &memcache.Item{
		Key:    key,
		Object: count + 1,
	})
	if err != nil {
		log.Infof(ctx, "counter %d", count)
	}

	version := appengine.VersionID(ctx)
	_, err = fmt.Fprintf(w, "Hello : %d, %s, %+v", count, version, time.Now())
	if err != nil {
		fmt.Printf("failed write to response. err=%v\n", err)
	}
}
