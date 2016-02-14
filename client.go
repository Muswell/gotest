// Gotest provides an http.Client that can be registered to hand RoundTrip requests
package gotest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Url is a string which represents a RoundTrip function. Just used for code clarity.
type roundTrip func(*http.Request) (*http.Response, error)

// RegisteredTransport is an http.RoundTripper which maps request urls and methods to a server.
type RegisteredTransport struct {
	// Register stores a map of the RoundTrip function to call for a url and method.
	register map[string]map[string]roundTrip
}

// Register adds a RoundTrip to the request registry.
func (tr RegisteredTransport) Register(url, method string, fn roundTrip) {
	method = strings.ToLower(method)
	// todo make concurrency safe.
	if tr.register[url] == nil {
		tr.register[url] = make(map[string]roundTrip)
	}
	tr.register[url][method] = fn
}

// UnRegister removes a RoundTrip from the request registry.
func (tr RegisteredTransport) UnRegister(url, method string) {
	method = strings.ToLower(method)
	// todo make concurrency safe.
	delete(tr.register[url], method)
}

func (tr RegisteredTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	method := strings.ToLower(r.Method)
	roundTrip, ok := tr.register[r.URL.Path][method]

	if !ok {
		return nil, fmt.Errorf("Url %s and Method %s are not registered.", r.URL.Path, method)
	}

	return roundTrip(r)
}

type RegisteredClient struct {
	*http.Client
}

func (rc *RegisteredClient) Register(url, method string, fn roundTrip) {
	switch t := rc.Transport.(type) {
	case RegisteredTransport:
		t.Register(url, method, fn)
		return
	default:
		log.Fatalf("Something went wrong. RegisteredClient not initialized correctly. %T\n", rc.Transport)
	}
}

func (rc *RegisteredClient) UnRegister(url, method string) {
	switch t := rc.Transport.(type) {
	case RegisteredTransport:
		t.UnRegister(url, method)
		return
	default:
		log.Fatalf("Something went wrong. RegisteredClient not initialized correctly. %T\n", rc.Transport)
	}
}

func NewRegisteredClient() *RegisteredClient {
	return &RegisteredClient{
		&http.Client{
			Transport: RegisteredTransport{register: make(map[string]map[string]roundTrip)},
		},
	}
}

// NopCloser is a Reader that satisfies the closer interface
type NopCloser struct {
	io.Reader
}

func (NopCloser) Close() error { return nil }
