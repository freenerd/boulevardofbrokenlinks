package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	SENDGRID_API_ENDPOINT = "https://api.sendgrid.com/api/mail.send.json"
)

func main() {
	checked := make(Checked)

	// setup handling of checks
	go func() {
		// wait until a check is done, if so handle it
		for {
			select {
			case downs := <-checked:
				handleDowns(downs)
			}
		}
	}()

	// start server that triggers checks
	http.HandleFunc("/trigger", triggerHandler(checkURL, checked))
	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}
	fmt.Printf("listening on port %s ...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleDowns(downs Downs) {
	if len(downs) < 1 {
		// nothing to do here
		return
	}

	ds := []Down{}

	// fetch all downs
loop:
	for {
		select {
		case d := <-downs:
			ds = append(ds, d)
		default:
			// drained channel, continue with processing
			break loop
		}
	}

	if MaySendEmail() {
		emailDowns(ds)
	} else {
		printDowns(ds)
	}
}

func printDowns(ds []Down) {
	for _, d := range ds {
		log.Println(d.String())
	}
}

func emailDowns(ds []Down) {
	sendEmail := sendEmailFunc(SENDGRID_API_ENDPOINT)
	sendEmail(ds[0].Origin, ds)
}
