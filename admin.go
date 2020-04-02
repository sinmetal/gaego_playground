package main

import (
	"fmt"
	"net/http"
	"time"
)

func adminHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Hello Admin: %+v", time.Now())
	if err != nil {
		fmt.Printf("failed write to response. err=%v\n", err)
	}
}
