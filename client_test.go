package gotest

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisteredClient(t *testing.T) {
	client := NewRegisteredClient()
	url := "http.google.com"
	if _, err := client.Get(url); err == nil {
		t.Errorf("Expected Get request to fail due to url not registered")
	}

	client.Register(url, "Get", func(req *http.Request) (*http.Response, error) {
		log.Println("Google RoundTripper called")
		w := httptest.NewRecorder()

		w.Write([]byte("Hello Google"))

		r := &http.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Header:        w.Header(),
			Body:          NopCloser{w.Body},
			ContentLength: int64(w.Body.Len()),
			Close:         true,
			Request:       req,
		}
		return r, nil
	})

	resp, err := client.Get(url)
	if err != nil {
		t.Errorf("Unexpected error after %s was registered %s", url, err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Could not read response body", err)
	}
	s := string(b)
	if s != "Hello Google" {
		t.Errorf("incorrect response body got %s expected %s", s, "Hello Google")
	}

	client.UnRegister(url, "Get")
	if _, err := client.Get(url); err == nil {
		t.Errorf("Expected Get request to fail due to unregistered url")
	}
}
