package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
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

func homepageHandler(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFiles("html/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, []string{})
}

func configGetHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q["id"] == nil || q["id"][0] == "" {
		http.Redirect(w, r, "/", 301)
		return
	}

	i, err := strconv.Atoi(q["id"][0])
	if err != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	u, err := GetUser(i)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", 301)
		return
	}

	t, err := template.ParseFiles("html/config.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, *u)
}
