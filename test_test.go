package webapp

import (
	"context"
	. "gopkg.in/check.v1"
	"net/http"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

func middlewareWriter(data string) Middleware {
	return func(next ContextHandler) ContextHandler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(data))
			next(ctx, rw, req)
		}
	}
}

func middlewareAbort() Middleware {
	return func(_ ContextHandler) ContextHandler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("abort"))
		}
	}
}

var finalHandler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("H"))
}
