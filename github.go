package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
  "encoding/json"

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

    token, err := c.getAccessToken(q["code"][0])
    if err != nil {
      log.Println("ERROR while getting access token", err)
      http.Redirect(w, r, "/", 301)
      return
    }

    user, err := c.getUser(token)
    if err != nil {
      log.Println(err)
      http.Redirect(w, r, "/", 301)
      return
    }

    email, err := c.getEmail(token)
    if err != nil {
      log.Println(err)
      http.Redirect(w, r, "/", 301)
      return
    }
    user.Email = email

    err = user.Save()
    if err != nil {
      log.Println(err)
      http.Redirect(w, r, "/", 301)
      return
    }
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

type githubUserJson struct {
  Login string `json:"login"`
  Id    int    `json:"id"`
}

func (c *githubClient) getUser(token string) (*githubUser, error) {
  body, err := c.authorizedCall("GET", token, "https://api.github.com/user")
  if err != nil {
    return nil, err
  }

  var user githubUserJson
  err = json.Unmarshal(body, &user)
  if err != nil {
    return nil, err
  }

  return &githubUser{
    AccessToken: token,
    GithubId: user.Id,
    Login: user.Login,
  }, nil
}

type githubEmailJson struct {
  Email string `json:"email"`
  Verified bool `json:"verified"`
  Primary bool `json:"primary"`
}

func (c *githubClient) getEmail(token string) (string, error) {
  body, err := c.authorizedCall("GET", token, "https://api.github.com/user/emails")
  if err != nil {
    return "", err
  }

  var emails []githubEmailJson
  err = json.Unmarshal(body, &emails)
  if err != nil {
    return "", err
  }

  if len(emails) == 0 {
    return "", fmt.Errorf("Expected to receive emails, got 0")
  }

  return emails[0].Email, nil
}

func (c *githubClient) authorizedCall(method, token, url string) ([]byte, error) {
  req, err := http.NewRequest(method, url, nil)
  if err != nil {
    return nil, err
  }
  req.SetBasicAuth(token, "x-oauth-basic")
  req.Header.Set("Accept", "application/vnd.github.v3.full+json")

  client := http.Client{}
  res, err := client.Do(req)
  if err != nil {
    return nil, err
  }

  body, err := ioutil.ReadAll(res.Body)
  res.Body.Close()
  if err != nil {
    return nil, err
  }

  return body, nil
}
