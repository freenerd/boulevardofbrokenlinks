package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	ENV = map[string]string{
		"SENDGRID_USERNAME": "sendgridusername",
		"SENDGRID_PASSWORD": "sendgridpassword",
		"EMAIL_RECIPIENT":   "email@example.com",
	}
)

func setupTestSendEmailHandler(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected send email to be called with POST, got %v", r.Method)
		}

		err := r.ParseForm()
		if err != nil {
			t.Errorf("couldn't parse form data")
			return
		}

		for _, field := range []struct {
			k string
			v string
		}{
			{"api_user", ENV["SENDGRID_USERNAME"]},
			{"api_key", ENV["SENDGRID_PASSWORD"]},
			{"to", ENV["EMAIL_RECIPIENT"]},
			{"toname", ENV["EMAIL_RECIPIENT"]},
			{"subject", "origin"},
			{"text", `
origin: 404 http://example.com/test`},
			{"from", ENV["EMAIL_RECIPIENT"]},
		} {
			p, present := r.PostForm[field.k]
			if !present {
				t.Errorf("expected post form key: %v with value: %v, got nothing", field.k, field.v)
				continue
			}

			// assumption: only one value per key
			if p[0] != field.v {
				t.Errorf("expected post form key: %v with value: %v, got value %v", field.k, field.v, p[0])
			}
		}

		w.WriteHeader(200)
	}
}

func TestSendEmail(t *testing.T) {
	setEnv()
	testHandler := setupTestSendEmailHandler(t)
	dummy := httptest.NewServer(http.HandlerFunc(testHandler))
	defer dummy.Close()

	f := sendEmailFunc("http://" + dummy.Listener.Addr().String())
	downs := []Down{
		{
			Origin: "origin",
			Url:    "http://example.com/test",
			Status: 404,
		},
	}
	resp, err := f("origin", downs)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("sendEmailFunc has not done http call")
	}
}

func TestMaySendEmail(t *testing.T) {
	os.Clearenv()

	if MaySendEmail() {
		t.Errorf("expected false if no env vars set")
	}

	setEnv()

	if !MaySendEmail() {
		t.Errorf("expected true with correct env vars set")
	}
}

func setEnv() {
	for k, v := range ENV {
		os.Setenv(k, v)
	}
}
