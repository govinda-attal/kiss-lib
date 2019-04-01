// Package httputil provides few utilities for http transport.
// This package may grow organically and later can be organised as desired.
// It provides:
//
// a) Wrapper handler that can be used to wrap application specific http handlers allowing simplified error handling.
//
// b) Authentication handler can be applied to a specific path & HTTP verb combination. Ideally this is to be used when say one or few HTTP verbs require authentication and others don't on the same resource path.
//
// c) Tracks unquiue X-Request-ID header field to its execution span. Wrapper handler will do this for you.
//
// d) Custom Not-Found (404) handler that returns HTTP 404 Status along with custom JSON message - {msg: "Not Found: Resource path not mapped"}
//
// e) JSON Request Binder and JSON Response Renderer utility methods to simplify JSON unmarshal and marshalling for HTTP request and responses. This will suit gorilla mux router/handler implementations.
//
// d) JSON & File Request Binder utility methods to simplify Multipart request methods.
//
// NOTE: Within Golang, it is an anti-pattern to dump utility functions to utility based packages. It is rather advised to organise them as per their purpose.
package httputil
