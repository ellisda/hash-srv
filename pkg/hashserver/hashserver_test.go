package hashserver

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var hashVal = [64]byte{0, 1, 2, 3, 4, 5, 6, 7}

func getTestServer() (*HashServer, *httptest.ResponseRecorder) {
	return NewHashServer(8080), httptest.NewRecorder()
}

func TestHashRequest(t *testing.T) {
	srv, recorder := getTestServer()

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
	srv, recorder := getTestServer()

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

func TestGetHashRequest(t *testing.T) {
	srv, recorder := getTestServer()

	req, err := http.NewRequest("GET", "/hash/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	http.HandlerFunc(srv.getHash).ServeHTTP(recorder, req)
	if recorder.Code != 400 {
		t.Error("Expected HTTP Response Status 400 before hash key exists")
	}

	recorder = httptest.NewRecorder()
	srv.hashes[1] = hashVal
	http.HandlerFunc(srv.getHash).ServeHTTP(recorder, req)
	if recorder.Code != 200 {
		t.Error("Expected HTTP Response Status 200 after hash key exists")
	}
	if recorder.Body.String() != base64.StdEncoding.EncodeToString(hashVal[:]) {
		t.Error("Expected HTTP Response to match base64 hashed value")
	}
}
