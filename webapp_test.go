package webapp

import (
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
)

type WebAppSuite struct{}

var _ = Suite(&WebAppSuite{})

func (s *WebAppSuite) TestNotFound(c *C) {
	app := New()
	app.NotFound(ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Not found"))
	}))

	response := doTestRequest(app, "GET", "/not/found/path")

	c.Assert(response.Code, Equals, 404)
	c.Assert(response.Body.String(), Equals, "Not found")
}

func (s *WebAppSuite) TestNotFoundWithGlobalMiddlewareHandlerIsSet(c *C) {
	app := New()
	called := false
	app.Use(func(next ContextHandler) ContextHandler {
		called = true
		return next
	})
	app.NotFound(ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Not found"))
	}))

	response := doTestRequest(app, "GET", "/not/found/path")

	c.Assert(called, Equals, true)
	c.Assert(response.Code, Equals, 404)
	c.Assert(response.Body.String(), Equals, "Not found")
}

func (s *WebAppSuite) TestNotFoundWithGlobalMiddlewareAfterHandlerIsSet(c *C) {
	app := New()
	app.NotFound(ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Not found"))
	}))
	called := false
	app.Use(func(next ContextHandler) ContextHandler {
		called = true
		return next
	})

	response := doTestRequest(app, "GET", "/not/found/path")

	c.Assert(called, Equals, true)
	c.Assert(response.Code, Equals, 404)
	c.Assert(response.Body.String(), Equals, "Not found")
}

func (s *WebAppSuite) TestNotAllowed(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))

	response := doTestRequest(app, "POST", "/test")

	c.Assert(response.Code, Equals, 405)
	c.Assert(response.Body.String(), Equals, "Method Not Allowed\n")
}

func (s *WebAppSuite) TestNotAllowedCustomFunction(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))
	app.MethodNotAllowed(ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		rw.Write([]byte("Method no no allowed"))
	}))

	response := doTestRequest(app, "POST", "/test")

	c.Assert(response.Code, Equals, 405)
	c.Assert(response.Body.String(), Equals, "Method no no allowed")
}

func (s *WebAppSuite) TestNotAllowedWithGlobalMiddlewareBeforeHandlerIsSet(c *C) {
	app := New()
	called := false
	app.Use(func(next ContextHandler) ContextHandler {
		called = true
		return next
	})
	app.GET("/test", ContextHandlerFunc(finalHandler))
	app.MethodNotAllowed(ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		rw.Write([]byte("Method no no allowed"))
	}))

	response := doTestRequest(app, "POST", "/test")

	c.Assert(called, Equals, true)
	c.Assert(response.Code, Equals, 405)
	c.Assert(response.Body.String(), Equals, "Method no no allowed")
}

func (s *WebAppSuite) TestNotAllowedWithGlobalMiddlewareAfterHandlerIsSet(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))
	app.MethodNotAllowed(ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		rw.Write([]byte("Method no no allowed"))
	}))
	called := false
	app.Use(func(next ContextHandler) ContextHandler {
		called = true
		return next
	})

	response := doTestRequest(app, "POST", "/test")

	c.Assert(called, Equals, true)
	c.Assert(response.Code, Equals, 405)
	c.Assert(response.Body.String(), Equals, "Method no no allowed")
}

func (s *WebAppSuite) TestNotFoundIsTriggerdWhenNotAllowedIsSetToNil(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))
	app.MethodNotAllowed(nil)

	response := doTestRequest(app, "POST", "/test")

	c.Assert(response.Code, Equals, 404)
}

func (s *WebAppSuite) TestHandleOptions(c *C) {
	app := New()
	app.HandleOptions(true)
	app.With(middlewareWriter("m1")).GET("/test", ContextHandlerFunc(finalHandler))

	response := doTestRequest(app, "OPTIONS", "/test")

	c.Assert(response.Code, Equals, 200)
	c.Assert(response.Body.String(), Equals, "")
}

func (s *WebAppSuite) TestNotHandleOptions(c *C) {
	app := New()
	app.HandleOptions(false)
	app.POST("/test", ContextHandlerFunc(finalHandler))

	response := doTestRequest(app, "OPTIONS", "/test")

	c.Assert(response.Code, Equals, 404)
}

func (s *WebAppSuite) TestRedirectTrailingSlash(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))

	app.RedirectTrailingSlash(true)
	response := doTestRequest(app, "GET", "/test/")

	c.Assert(response.Code, Equals, 301)
	c.Assert(response.Header().Get("Location"), Equals, "/test")
}

func (s *WebAppSuite) TestRedirectTrailingSlashFalse(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))

	app.RedirectTrailingSlash(false)
	response := doTestRequest(app, "GET", "/test/")

	c.Assert(response.Code, Equals, 404)
}

func (s *WebAppSuite) TestRedirectFixedPath(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))

	app.RedirectFixedPath(true)
	response := doTestRequest(app, "GET", "/abc/../test")

	c.Assert(response.Code, Equals, 301)
	c.Assert(response.Header().Get("Location"), Equals, "/test")
}

func (s *WebAppSuite) TestRedirectFixedPathFalse(c *C) {
	app := New()
	app.GET("/test", ContextHandlerFunc(finalHandler))

	app.RedirectFixedPath(false)
	response := doTestRequest(app, "GET", "/abc/../test")

	c.Assert(response.Code, Equals, 404)
}
