package main

import (
	"testing"

	"fmt"
	"net/http"
	"net/http/httptest"
)

func TestTriggerHandler(t *testing.T) {
	expectedBody := "ok"
	triggerHandlerTriggered := false
	handler := triggerHandler(func(string, chan Down) error {
		triggerHandlerTriggered = true
    return nil
	}, make(chan Down))
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

	if !triggerHandlerTriggered {
		t.Errorf("callback was not triggered")
	}
}
