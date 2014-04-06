package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"log"
)

type githubClient struct {
	client_id     string
	client_secret string
	redirect_uri  string
	scope         string
}

func (c *githubClient) authorizeURL() string {
	u, err := url.Parse("https://github.com/login/oauth/authorize")
	if err != nil {
		log.Fatal("Malformed Github URL", err)
	}

	q := u.Query()
	q.Set("client_id", c.client_id)
	q.Set("redirect_uri", c.redirect_uri)
	q.Set("scope", c.scope)
	u.RawQuery = q.Encode()

	return u.String()
}

func (c *githubClient) authorizeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, c.authorizeURL(), 301)
	}
}

func (c *githubClient) callbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if q["code"] == nil || q["code"][0] == "" {
			http.Redirect(w, r, c.authorizeURL(), 301)
			return
		}

		fmt.Println(c.getAccessToken(q["code"][0]))
	}
}

func (c *githubClient) accessTokenURL(code string) url.URL {
	u, err := url.Parse("https://github.com/login/oauth/access_token")
	if err != nil {
		log.Fatal("Malformed Github URL", err)
	}

	q := u.Query()
	q.Set("client_id", c.client_id)
	q.Set("client_secret", c.client_secret)
	q.Set("code", code)
	u.RawQuery = q.Encode()

	return *u
}

func (c *githubClient) getAccessToken(code string) (string, error) {
	u := c.accessTokenURL(code)
	hc := http.Client{}

	res, err := hc.PostForm(u.String(), u.Query())
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	pb, err := url.ParseQuery(string(b[:]))
	if err != nil {
		return "", err
	}

	if pb["scope"] == nil {
		return "", fmt.Errorf("Could not fetch scope")
	}
	scope := pb["scope"][0]
	if scope != c.scope {
		return "", fmt.Errorf("Incorrect scope. Expected: %s got: %s", c.scope, scope)
	}

	if pb["access_token"] == nil {
		return "", fmt.Errorf("Could not fetch access token")
	}
	token := pb["access_token"][0]
	if token == "" {
		return "", fmt.Errorf("Could not fetch access token")
	}

	return token, nil
}
