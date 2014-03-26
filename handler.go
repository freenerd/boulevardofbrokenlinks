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

func triggerHandler(fn func(string, Checked) error, checked Checked) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// respond
		fmt.Fprintf(w, "ok")

		// start processing
		go func() {
			err := fn(CHECK_URL, checked)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
