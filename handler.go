package webapp

// Package webapp handler provides a bridge between http.Handler and net/context.
// for the routing the router from github.com/julienschmidt/httprouter is used

import (
	"golang.org/x/net/context"
	"net/http"
)

type ContextHandler func(context.Context, http.ResponseWriter, *http.Request)
