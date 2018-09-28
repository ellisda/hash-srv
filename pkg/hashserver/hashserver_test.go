package hashserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var hashVal = [64]byte{0, 1, 2, 3, 4, 5, 6, 7}

func getTestServer() *HashServer {
	return NewHashServer(8080)
}

func execHandler(t *testing.T, h func(http.ResponseWriter, *http.Request), req *http.Request, expectedStatus int,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	http.HandlerFunc(h).ServeHTTP(recorder, req)
	if recorder.Code != expectedStatus {
		t.Fatalf("Expected http.Handler to return Status %d, got %d", expectedStatus, recorder.Code)
	}
	return recorder
}

func TestHashRequest(t *testing.T) {
	srv := getTestServer()

	req, err := http.NewRequest("POST", "/hash", strings.NewReader("password=foo"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		t.Fatal(err)
	}
	recorder := execHandler(t, srv.hashRequest, req, 202)
	body := recorder.Body.String()
	if body != "1" {
		t.Fatal("Expected response body to start at 1")
	}

	recorder = execHandler(t, srv.hashRequest, req, 202)
	body = recorder.Body.String()
	if body != "2" {
		t.Fatalf("Expected response body to continue at 2")
	}
}

func TestHashRequestRejected(t *testing.T) {
	srv := getTestServer()

	emptyReq, err := http.NewRequest("POST", "/hash", nil)
	if err != nil {
		t.Fatal(err)
	}
	execHandler(t, srv.hashRequest, emptyReq, 400)

	badReq, err := http.NewRequest("POST", "/hash", strings.NewReader("foo=bar"))
	badReq.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		t.Fatal(err)
	}
	execHandler(t, srv.hashRequest, badReq, 400)
}

func TestGetHashRequest(t *testing.T) {
	srv := getTestServer()

	req, err := http.NewRequest("GET", "/hash/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	execHandler(t, srv.getHash, req, 400)

	srv.hashes[1] = hashVal
	recorder := execHandler(t, srv.getHash, req, 200)
	if recorder.Body.String() != base64.StdEncoding.EncodeToString(hashVal[:]) {
		t.Error("Expected HTTP Response to match base64 hashed value")
	}
}

func TestGetStats(t *testing.T) {
	srv := getTestServer()

	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := execHandler(t, srv.getStats, req, 200)
	if recorder.Body.String() != "{\"total\":0,\"average\":0}" {
		t.Error("Stats produced unexpected content")
	}

	srv.stats.NumProcessed = 1
	recorder = execHandler(t, srv.getStats, req, 200)
	var stats requestStats
	json.Unmarshal(recorder.Body.Bytes(), &stats)
	if stats.NumProcessed != 1 {
		t.Error("Stats produced unexpected content")
	}
}
