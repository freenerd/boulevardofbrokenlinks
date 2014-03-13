package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

// returns a function that sends an email.
// address is the address of the endpoint to send email to.
func sendEmailFunc(address string) func(string, []Down) (*http.Response, error) {
	return func(origin string, downs []Down) (*http.Response, error) {
		text := ""
		for _, d := range downs {
			text = fmt.Sprintf("%s\n%s", text, d.String())
		}

		v := url.Values{}
		v.Add("api_user", os.Getenv("SENDGRID_USERNAME"))
		v.Add("api_key", os.Getenv("SENDGRID_PASSWORD"))
		v.Add("to", os.Getenv("EMAIL_RECIPIENT"))
		v.Add("toname", os.Getenv("EMAIL_RECIPIENT"))
		v.Add("subject", origin)
		v.Add("text", text)
		v.Add("from", os.Getenv("EMAIL_RECIPIENT"))

		resp, err := http.PostForm(address, v)
		log.Println("Email sent")
		log.Println(resp)
		log.Println(err)

		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func MaySendEmail() bool {
	return os.Getenv("SENDGRID_USERNAME") != "" &&
		os.Getenv("SENDGRID_PASSWORD") != "" &&
		os.Getenv("EMAIL_RECIPIENT") != ""
}
