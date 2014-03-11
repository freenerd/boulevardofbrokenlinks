package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var CHECK_URL string

func init() {
	CHECK_URL = os.Getenv("CHECK_URL")

	if CHECK_URL == "" {
		log.Println("CHECK_URL env var not set")
	}
}

func triggerHandler(fn func(string)) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "ok")
		fn(CHECK_URL)
	}
}
