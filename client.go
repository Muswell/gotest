// Gotest provides an http.Client that can be registered to hand RoundTrip requests
package gotest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Url is a string which represents a RoundTrip function. Just used for code clarity.
type RoundTrip func(*http.Request) (*http.Response, error)

// RegisteredTransport is an http.RoundTripper which maps request urls and methods to a server.
type RegisteredTransport struct {
	// Register stores a map of the RoundTrip function to call for a url and method.
	register map[string]map[string]RoundTrip
}

// Register adds a RoundTrip to the request registry.
func (tr RegisteredTransport) Register(url, method string, fn RoundTrip) {
	method = strings.ToLower(method)
	// todo make concurrency safe.
	if tr.register[url] == nil {
		tr.register[url] = make(map[string]RoundTrip)
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

func (rc *RegisteredClient) Register(url, method string, fn RoundTrip) {
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
			Transport: RegisteredTransport{register: make(map[string]map[string]RoundTrip)},
		},
	}
}

// NopCloser is a Reader that satisfies the closer interface
type NopCloser struct {
	io.Reader
}

func (NopCloser) Close() error { return nil }

func NewSimpleRoundTrip(body []byte, headers map[string]string) RoundTrip {
	return func(req *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()

		w.Write(body)

		for k, v := range headers {
			w.Header().Set(k, v)
		}

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
	}
}
