package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serverWithStatus(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func TestCheckUrl(t *testing.T) {
	checked := make(Checked, 100)
	var downs Downs

	// incorrect url
	err := checkURL("incorrecturl", checked)
	if err == nil {
		t.Error("expected incorrect url to throw error")
	}
	drainChan(checked)

	// empty response
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	dummy := httptest.NewServer(emptyHandler)
	err = checkURL("http://"+dummy.Listener.Addr().String(), checked)
	if err != nil {
		t.Error("expected correct URL to not throw error, got ", err)
	}
	if len(checked) != 1 {
		t.Errorf("expected checked to have one element after executed checkURL, got %d elements", len(checked))
		return // return since next call blocks
	}
	downs = <-checked
	if len(downs) > 0 {
		t.Errorf("expected empty response to not populate chan, got %d items", len(checked))
	}
	drainChan(checked)
	dummy.Close()

	// malformed HTML
	malformedHTMLHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<<><><>>")
	})
	dummy = httptest.NewServer(malformedHTMLHandler)
	err = checkURL("http://"+dummy.Listener.Addr().String(), checked)
	if err != nil {
		t.Error("expected correct URL to not throw error, got ", err)
	}
	if len(checked) != 1 {
		t.Errorf("expected checked to have one element after executed checkURL, got %d elements", len(checked))
		return // return since next call blocks
	}
	downs = <-checked
	if len(downs) > 0 {
		t.Errorf("expected malformed response to not populate chan, got %d items", len(downs))
	}
	drainChan(checked)
	dummy.Close()

	// correct responses
	servers := []*httptest.Server{
		serverWithStatus(200),
		serverWithStatus(301),
		serverWithStatus(400),
		serverWithStatus(404),
		serverWithStatus(500),
		serverWithStatus(503),
	}
	var body string
	for _, s := range servers {
		body = fmt.Sprintf("%v\n<a href=\"http://%v\" />", body, s.Listener.Addr().String())
	}
	HTMLHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})
	dummy = httptest.NewServer(HTMLHandler)
	err = checkURL("http://"+dummy.Listener.Addr().String(), checked)
	if err != nil {
		t.Error("expected correct URL to not throw error, got ", err)
	}
	if len(checked) != 1 {
		t.Errorf("expected checked to have one element after executed checkURL, got %d elements", len(checked))
		return // return since next call blocks
	}
	downs = <-checked
	if len(downs) != 4 {
		t.Errorf("expected response to populate chan with 4 items, got %d items", len(checked))
	}
	for _, s := range servers {
		s.Close()
	}
}

func drainChan(c Checked) {
	for {
		if len(c) == 0 {
			return
		}
		<-c
	}
}
