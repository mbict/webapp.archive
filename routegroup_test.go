package webapp

import (
	"bytes"
	"github.com/julienschmidt/httprouter"
	. "gopkg.in/check.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
)

type RouteGroupSuite struct{}

var _ = Suite(&RouteGroupSuite{})

func doTestRequest(rg http.Handler, method, path string) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	req.RemoteAddr = "123.123.123.123"
	rg.ServeHTTP(rw, req)
	return rw
}

func (s *RouteGroupSuite) TestRouteGroup(c *C) {

	tests := []struct {
		method   string
		path     string
		response string
		prepare  func(RouteGroup)
	}{
		{
			//root test
			method:   "GET",
			path:     "/",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.GET("/", finalHandler)
			},
		}, {
			//root test
			method:   "GET",
			path:     "/",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.GET("", finalHandler)
			},
		}, {
			//strange routes patterns with closing slash
			method:   "GET",
			path:     "/test/",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("test").GET("/", finalHandler)
			},
		}, {
			//strange routes patterns without closing slash
			method:   "GET",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("test").GET("", finalHandler)
			},
		}, {
			//root test
			method:   "GET",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.GET("/test", finalHandler)
			},
		}, {
			//root test with middleware
			method:   "GET",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).GET("/test", finalHandler)
			},
		}, {
			//group test
			method:   "GET",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").GET("/test", finalHandler)
			},
		}, {
			//group with middleware
			method:   "GET",
			path:     "/group/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group", middlewareWriter("M1")).GET("/test", finalHandler)
			},
		}, {
			method:   "GET",
			path:     "/group/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).Group("/group").GET("/test", finalHandler)
			},
		}, {
			method:   "GET",
			path:     "/group/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").With(middlewareWriter("M1")).GET("/test", finalHandler)
			},
		}, {
			method:   "GET",
			path:     "/group/test",
			response: "M1M2M3H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).Group("/group", middlewareWriter("M2")).With(middlewareWriter("M3")).GET("/test", finalHandler)
			},
		}, {
			method:   "GET",
			path:     "/group/group2/test",
			response: "M1M2H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group", middlewareWriter("M1")).Group("/group2", middlewareWriter("M2")).GET("/test", finalHandler)
			},
		},

		//test POST
		{
			method:   "POST",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.POST("/test", finalHandler)
			},
		}, {
			method:   "POST",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).POST("/test", finalHandler)
			},
		}, {
			method:   "POST",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").POST("/test", finalHandler)
			},
		},

		//test PUT
		{
			method:   "PUT",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.PUT("/test", finalHandler)
			},
		}, {
			method:   "PUT",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).PUT("/test", finalHandler)
			},
		}, {
			method:   "PUT",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").PUT("/test", finalHandler)
			},
		},

		//test PATCH
		{
			method:   "PATCH",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.PATCH("/test", finalHandler)
			},
		}, {
			method:   "PATCH",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).PATCH("/test", finalHandler)
			},
		}, {
			method:   "PATCH",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").PATCH("/test", finalHandler)
			},
		},

		//test DELETE
		{
			method:   "DELETE",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.DELETE("/test", finalHandler)
			},
		}, {
			method:   "DELETE",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).DELETE("/test", finalHandler)
			},
		}, {
			method:   "DELETE",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").DELETE("/test", finalHandler)
			},
		},

		//test OPTIONS
		{
			method:   "OPTIONS",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.OPTIONS("/test", finalHandler)
			},
		}, {
			method:   "OPTIONS",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).OPTIONS("/test", finalHandler)
			},
		}, {
			method:   "OPTIONS",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").OPTIONS("/test", finalHandler)
			},
		},

		//test HEAD
		{
			method:   "HEAD",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.HEAD("/test", finalHandler)
			},
		}, {
			method:   "HEAD",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).HEAD("/test", finalHandler)
			},
		}, {
			method:   "HEAD",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").HEAD("/test", finalHandler)
			},
		},

		//test LINK
		{
			method:   "LINK",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.LINK("/test", finalHandler)
			},
		}, {
			method:   "LINK",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).LINK("/test", finalHandler)
			},
		}, {
			method:   "LINK",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").LINK("/test", finalHandler)
			},
		},

		//test UNLINK
		{
			method:   "UNLINK",
			path:     "/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.UNLINK("/test", finalHandler)
			},
		}, {
			method:   "UNLINK",
			path:     "/test",
			response: "M1H",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).UNLINK("/test", finalHandler)
			},
		}, {
			method:   "UNLINK",
			path:     "/group/test",
			response: "H",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").UNLINK("/test", finalHandler)
			},
		},

		//test serve static file
		{
			method:   "GET",
			path:     "/test.txt",
			response: "test file",
			prepare: func(rg RouteGroup) {
				rg.StaticFile("/test.txt", "./_test/file.txt")
			},
		}, {
			method:   "GET",
			path:     "/test/test.txt",
			response: "M1test file",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).StaticFile("/test/test.txt", "./_test/file.txt")
			},
		}, {
			method:   "GET",
			path:     "/group/test.txt",
			response: "test file",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").StaticFile("/test.txt", "./_test/file.txt")
			},
		},

		//test serve static directory
		{
			method:   "GET",
			path:     "/file.txt",
			response: "test file",
			prepare: func(rg RouteGroup) {
				rg.Static("/", "./_test")
			},
		}, {
			method:   "GET",
			path:     "/file.txt",
			response: "test file",
			prepare: func(rg RouteGroup) {
				rg.Static("", "./_test")
			},
		}, {
			method:   "GET",
			path:     "/test/file.txt",
			response: "M1test file",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).Static("/test", "./_test")
			},
		}, {
			method:   "GET",
			path:     "/test/file.txt",
			response: "M1test file",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("M1")).Static("/test/", "./_test")
			},
		}, {
			method:   "GET",
			path:     "/group/file.txt",
			response: "test file",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").Static("/", "./_test")
			},
		}, {
			method:   "GET",
			path:     "/group/file.txt",
			response: "test file",
			prepare: func(rg RouteGroup) {
				rg.Group("/group").Static("", "./_test")
			},
		},
	}

	for index, test := range tests {
		r := httprouter.New()
		rg := newRouteGroup(r)
		test.prepare(rg)
		response := doTestRequest(rg, test.method, test.path)

		c.Check(response.Code, Equals, 200, Commentf("test %d failed status code %d", index, response.Code))
		c.Check(response.Body.String(), Equals, test.response, Commentf("test %d failed (method) %s (path) %s", index, test.method, test.path))
	}
}

func (s *RouteGroupSuite) TestWithCreatesNewGroup(c *C) {
	rg := newRouteGroup(nil)

	rguse := rg.With()

	c.Assert(reflect.ValueOf(rg).Elem().UnsafeAddr(), Not(Equals), reflect.ValueOf(rguse).Elem().UnsafeAddr())
}

func (s *RouteGroupSuite) TestUseAppendMiddlewareGroup(c *C) {
	rg := newRouteGroup(nil).(*routeGroup)

	rg.Use(middlewareWriter("a"), middlewareWriter("b"))

	c.Assert(rg.middleware, HasLen, 2)
}

func (s *RouteGroupSuite) TestRouteLogger(c *C) {

	tests := []struct {
		expected string
		prepare  func(RouteGroup)
	}{
		{
			expected: "^DELETE.*/trash.*-->.*github.com/mbict/webapp.glob.func.*\\(0 middlewares\\)\\n$",
			prepare: func(rg RouteGroup) {
				rg.DELETE("/trash", finalHandler)
			},
		}, {
			expected: "^POST.*/test/abc.*-->.*github.com/mbict/webapp.middlewareWriter.func.*\\(1 middlewares\\)\\n$",
			prepare: func(rg RouteGroup) {
				rg.With(middlewareWriter("a")).POST("/test/abc", finalHandler)
			},
		}, {
			expected: "^GET.*/hi/world.*-->.*github.com/mbict/webapp.glob.func.*\\(0 middlewares\\)\\n$",
			prepare: func(rg RouteGroup) {
				rg.Group("/hi").GET("/world", finalHandler)
			},
		},
	}

	for index, test := range tests {
		sw := bytes.NewBuffer([]byte{})
		logger := log.New(sw, "", 0)
		rg := newRouteGroup(httprouter.New())
		rg.Logger(logger)
		regxp := regexp.MustCompile(test.expected)

		test.prepare(rg)

		c.Check(regxp.MatchString(sw.String()), Equals, true, Commentf("test %d failed", index))
	}
}
