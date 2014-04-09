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

var (
	db dbType
)

func main() {
  fmt.Println(newCheckToken())
  fmt.Println(newCheckToken())
  fmt.Println(newCheckToken())

	// setup database connection
	db = dbType{url: os.Getenv("DATABASE_URL")}
	if err := db.connect(); err != nil {
		log.Fatal(err)
	}
	db.Check()

	// setup handling of checks
	checked := make(Checked)
	go func() {
		// wait until a check is done, if so handle it
		for {
			select {
			case downs := <-checked:
				handleDowns(downs)
			}
		}
	}()

	// setup github connect flow
	gh := githubClient{
		client_id:     os.Getenv("GITHUB_CLIENT_ID"),
		client_secret: os.Getenv("GITHUB_CLIENT_SECRET"),
		redirect_uri:  os.Getenv("GITHUB_CALLBACK"),
		scope:         "user:email",
	}
	http.HandleFunc("/login/github/authorize", gh.authorizeHandler())
	http.HandleFunc("/login/github/callback", gh.callbackHandler())

	// setup user-facing websites
	http.HandleFunc("/", homepageHandler)
	http.HandleFunc("/config", configGetHandler)

	// setup trigger
	http.HandleFunc("/trigger", triggerHandler(checkURL, checked))
	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	// start server
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
