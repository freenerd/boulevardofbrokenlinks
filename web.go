package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	SENDGRID_API_ENDPOINT = "https://api.sendgrid.com/api/mail.send.json"
)

var (
	downs = make(chan Down)
)

func main() {
	// start aggregate routine
	aggregates()

	// start server that triggers checks
	http.HandleFunc("/trigger", triggerHandler(checkURL, downs))
	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}
	fmt.Printf("listening on port %s ...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func aggregates() {
	go func() {
		for {
			collectAggregates()
			time.Sleep(1 * time.Minute)
		}
	}()
}

func collectAggregates() {
	ds := Aggregates{}

loop:
	for {
		select {
		case d := <-downs:
			ds[d.Origin] = append(ds[d.Origin], d)
		default:
			// drained channel, continue with processing
			break loop
		}
	}

	if MaySendEmail() {
		emailAggregates(ds)
	} else {
		printAggregates(ds)
	}
}

func printAggregates(ds Aggregates) {
	for _, downs := range ds {
		for _, d := range downs {
			log.Println(d.String())
		}
	}
}

func emailAggregates(ds Aggregates) {
	sendEmail := sendEmailFunc(SENDGRID_API_ENDPOINT)
	for origin, downs := range ds {
		sendEmail(origin, downs)
	}
}
