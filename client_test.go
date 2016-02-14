package gotest

import (
	"io/ioutil"
	"testing"
)

func TestRegisteredClient(t *testing.T) {
	client := NewRegisteredClient()
	url := "http.google.com"
	if _, err := client.Get(url); err == nil {
		t.Errorf("Expected Get request to fail due to url not registered")
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "text/html"

	client.Register(url, "Get", NewSimpleRoundTrip([]byte("Hello Google"), headers))

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

	if resp.Header.Get("Content-Type") != "text/html" {
		t.Errorf("incorrect response Content-Type got %s expected %s", resp.Header.Get("Content-Type"), "text/html")
	}

	client.UnRegister(url, "Get")
	if _, err := client.Get(url); err == nil {
		t.Errorf("Expected Get request to fail due to unregistered url")
	}
}
