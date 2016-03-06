package webapp

// Package webapp handler provides a bridge between http.Handler and net/context.
// for the routing the router from github.com/julienschmidt/httprouter is used

import (
	"golang.org/x/net/context"
	"net/http"
)

// ContextHandler is a net/context aware http.Handler
type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (h ContextHandlerFunc) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	h(ctx, rw, req)
}
