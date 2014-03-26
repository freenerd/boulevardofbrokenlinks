package main

import (
	"testing"

	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func TestTriggerHandler(t *testing.T) {
	expectedBody := "ok"
	checked := make(Checked, 1)
	handler := triggerHandler(func(string, Checked) error {
		checked <- make(Downs)
		return nil
	}, checked)
	recorder := httptest.NewRecorder()
	url := fmt.Sprintf("http://example.com/trigger")
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		t.Errorf("%s", err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Body.String() != expectedBody {
		t.Errorf("expected: %s. got: %s", expectedBody, recorder.Body.String())
	}

	tick := time.Tick(3 * time.Second)
	for {
		select {
		case <-tick:
			t.Error("callback was not triggered in time")
			return
		case <-checked:
			return
		}
	}
}
