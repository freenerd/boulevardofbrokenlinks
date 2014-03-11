package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go.net/html"
)

const (
	SENDGRID_API_ENDPOINT = "https://api.sendgrid.com/api/mail.send.json"
)

var (
	downs = make(chan down)
)

func main() {
	// start aggregate routine
	aggregates()

	// start server that triggers checks
	http.HandleFunc("/trigger", triggerHandler(checkURL))
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

func checkURL(origin string) {
	// parse incoming url
	up, err := url.Parse(origin)
	if err != nil {
		log.Println(err)
		return
	}
	SCHEME := up.Scheme
	HOST := up.Host

	resp, err := http.Get(origin)
	if err != nil {
		log.Println(err)
		return
	}

	tree, err := h5.New(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tree.Walk(func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					u, err := url.Parse(attr.Val)
					if err != nil {
						log.Println(err)
						continue
					}

					if !u.IsAbs() {
						// if its relative, we assume it's relative from the current path
						u.Scheme = SCHEME
						u.Host = HOST
					}

					checkHead(origin, u.String())
				}
			}
		}
	})
}

func checkHead(origin, url string) {
	go func() {
		resp, err := http.Head(url)
		if err != nil {
			// assuming problems with network or malformed url
			// ignore
			return
		}

		if resp.StatusCode >= 400 {
			// url is in trouble
			// some sites like amazon or google app engine don't like HEAD, let's retry with GET
			checkGet(origin, url)
		}
	}()
}

func checkGet(origin, url string) {
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			// assuming problems with network or malformed url
			// ignore
			return
		}

		if resp.StatusCode >= 400 {
			// url is down, down, down
			downs <- down{
				Origin: origin,
				Url:    url,
				Status: resp.StatusCode,
			}
		}
	}()
}
