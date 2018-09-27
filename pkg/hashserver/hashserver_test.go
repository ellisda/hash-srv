package hashserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHashRequest(t *testing.T) {
	srv := New(8080)

	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/hash", strings.NewReader("password=foo"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		t.Fatal(err)
	}

	http.HandlerFunc(srv.hashRequest).ServeHTTP(recorder, req)
	if recorder.Code != 202 {
		t.Fatal("Expected 202 Status")
	}
	body := recorder.Body.String()
	if body != "1" {
		t.Fatal("Expected response body to start at 1")
	}

	recorder = httptest.NewRecorder()
	http.HandlerFunc(srv.hashRequest).ServeHTTP(recorder, req)
	if recorder.Code != 202 {
		t.Fatal("Expected 202 Status")
	}
	body = recorder.Body.String()
	if body != "2" {
		t.Fatalf("Expected response body to continue at 2")
	}
}

func TestHashRequestRejected(t *testing.T) {
	srv := New(8080)
	recorder := httptest.NewRecorder()

	emptyReq, err := http.NewRequest("POST", "/hash", nil)
	if err != nil {
		t.Fatal(err)
	}

	http.HandlerFunc(srv.hashRequest).ServeHTTP(recorder, emptyReq)
	if recorder.Code != 400 {
		t.Fatal("Expected 400 Status for empty POST body")
	}

	badReq, err := http.NewRequest("POST", "/hash", strings.NewReader("foo=bar"))
	badReq.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		t.Fatal(err)
	}

	http.HandlerFunc(srv.hashRequest).ServeHTTP(recorder, badReq)
	if recorder.Code != 400 {
		t.Fatal("Expected 400 Status for POST body without 'password' form value")
	}
}
