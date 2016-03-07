package webapp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"time"
)

type MiddlewareSuite struct{}

var _ = Suite(&MiddlewareSuite{})

/******************************
 Recovery middleware
*******************************/

func panickingHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	panic("omg omg what a panic")
}

func (s *MiddlewareSuite) TestRecovery(c *C) {
	rg := newRouteGroup(httprouter.New())
	rg.With(Recovery(nil)).GET("/panic/test", panickingHandler)

	response := doTestRequest(rg, "GET", "/panic/test")

	c.Assert(response.Code, Equals, 500)
	c.Assert(response.Body.String(), Equals, "Internal server error")
}

func (s *MiddlewareSuite) TestRecoveryWithHandler(c *C) {
	var err error
	var errStackTrace []byte
	recoveryHandler := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		errStackTrace = ErrorStackTrace(ctx)
		err = Error(ctx)

		rw.WriteHeader(500)
		rw.Write([]byte("recovery handler"))
	}
	rg := newRouteGroup(httprouter.New())
	rg.With(Recovery(recoveryHandler)).GET("/panic/test", panickingHandler)

	response := doTestRequest(rg, "GET", "/panic/test")

	c.Assert(response.Code, Equals, 500)
	c.Assert(response.Body.String(), Equals, "recovery handler")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "omg omg what a panic")
	c.Assert(errStackTrace, Not(Equals), "")
}

func (s *MiddlewareSuite) TestError(c *C) {
	err := errors.New("test")
	ctx := context.WithValue(context.Background(), errorKey, err)

	c.Assert(Error(ctx), Equals, err)
}

func (s *MiddlewareSuite) TestErrorWithNilContext(c *C) {
	c.Assert(Error(nil), IsNil)
}

func (s *MiddlewareSuite) TestErrorOnInCompatibleContextValue(c *C) {
	ctx := context.WithValue(context.Background(), errorKey, int64(123))

	c.Assert(Error(ctx), IsNil)
}

func (s *MiddlewareSuite) TestErrorStackTrace(c *C) {
	stackTrace := []byte("test")
	ctx := context.WithValue(context.Background(), stackTraceKey, stackTrace)

	c.Assert(ErrorStackTrace(ctx), DeepEquals, stackTrace)
}

func (s *MiddlewareSuite) TestErrorStackTraceWithNilContext(c *C) {
	c.Assert(ErrorStackTrace(nil), IsNil)
}

func (s *MiddlewareSuite) TestErrorStackTraceOnInCompatibleContextValue(c *C) {
	ctx := context.WithValue(context.Background(), stackTraceKey, int64(123))

	c.Assert(ErrorStackTrace(ctx), IsNil)
}

/******************************
 Timeout middleware
*******************************/

func (s *MiddlewareSuite) TestTimeout(c *C) {
	rg := newRouteGroup(httprouter.New())
	rg.With(Timeout(5*time.Millisecond)).GET("/timeout", func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		time.Sleep(20 * time.Millisecond)
	})

	response := doTestRequest(rg, "GET", "/timeout")

	c.Assert(response.Code, Equals, 503)
	c.Assert(response.Body.String(), Equals, "Server request timeout")
}

func (s *MiddlewareSuite) TestTimeoutSuccessful(c *C) {
	rg := newRouteGroup(httprouter.New())
	rg.With(Timeout(30*time.Second)).GET("/timeout", func(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(200)
		time.Sleep(20 * time.Millisecond)
	})

	response := doTestRequest(rg, "GET", "/timeout")

	c.Assert(response.Code, Equals, 200)
}

/******************************
 UniqueRequestId middleware
*******************************/
func (s *MiddlewareSuite) TestUniqueRequestID(c *C) {
	requestId := ""
	rg := newRouteGroup(httprouter.New())
	rg.With(UniqueRequestID()).GET("/test", func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) {
		requestId = RequestID(ctx)
	})
	rgpat := regexp.MustCompile("^[^-]+-\\d+$")

	doTestRequest(rg, "GET", "/test")

	c.Assert(rgpat.MatchString(requestId), Equals, true)
}

func (s *MiddlewareSuite) TestUniqueRequestIDWithHeader(c *C) {
	requestId := ""
	rg := newRouteGroup(httprouter.New())
	rg.With(UniqueRequestID()).GET("/test", func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) {
		requestId = RequestID(ctx)
	})
	rw := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Add(RequestIDHeader, "test-12345")
	rg.ServeHTTP(rw, req)

	c.Assert(requestId, Equals, "test-12345")
}

func (s *MiddlewareSuite) TestRquestID(c *C) {
	ctx := context.WithValue(context.Background(), requestIDKey, "test")

	c.Assert(RequestID(ctx), DeepEquals, "test")
}

func (s *MiddlewareSuite) TestRequestIDWithNilContext(c *C) {
	c.Assert(RequestID(nil), Equals, "")
}

func (s *MiddlewareSuite) TestRequestIDOnInCompatibleContextValue(c *C) {
	ctx := context.WithValue(context.Background(), requestIDKey, int64(123))

	c.Assert(RequestID(ctx), Equals, "")
}

/******************************
 LogRequest middleware
*******************************/
func (s *MiddlewareSuite) TestLogRequest(c *C) {

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{
			method: "GET",
			path:   "/test",
			code:   200,
		}, {
			method: "POST",
			path:   "/test",
			code:   200,
		}, {
			method: "PUT",
			path:   "/test",
			code:   200,
		}, {
			method: "DELETE",
			path:   "/test",
			code:   200,
		}, {
			method: "HEAD",
			path:   "/test",
			code:   200,
		}, {
			method: "LINK",
			path:   "/test",
			code:   200,
		}, {
			method: "PATCH",
			path:   "/test",
			code:   200,
		}, {
			method: "OPTIONS",
			path:   "/test",
			code:   200,
		}, {
			method: "UNLINK",
			path:   "/test",
			code:   200,
		}, {
			//not found handler
			method: "GET",
			path:   "/notfound",
			code:   404,
		}, {
			//recovery middleware
			method: "GET",
			path:   "/internalservererror",
			code:   500,
		}, {
			//not allowed
			method: "POST",
			path:   "/test/notallowed",
			code:   405,
		}, {
			method: "GET",
			path:   "/redirectme",
			code:   301,
		},
	}

	iseh := func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		panic("panda ikes")
	}
	rh := func(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("location", "/test")
		rw.WriteHeader(http.StatusMovedPermanently)
	}

	for index, test := range tests {
		sw := bytes.NewBuffer([]byte{})
		logger := log.New(sw, "", 0)

		app := New()
		app.Use(LogRequest(logger), Recovery(nil))
		app.GET("/test", finalHandler)
		app.GET("/redirectme", rh)
		app.GET("/test/notallowed", finalHandler)
		app.POST("/test", finalHandler)
		app.PUT("/test", finalHandler)
		app.DELETE("/test", finalHandler)
		app.HEAD("/test", finalHandler)
		app.LINK("/test", finalHandler)
		app.PATCH("/test", finalHandler)
		app.OPTIONS("/test", finalHandler)
		app.UNLINK("/test", finalHandler)
		app.GET("/internalservererror", iseh)

		pattern := fmt.Sprintf("^\\d{4}/\\d{2}/\\d{2} - \\d{2}:\\d{2}:\\d{2} \\|.* %d .*\\|\\s+\\d+(\\.\\d+)?[\\wµ]?s \\|\\s+123.123.123.123 \\|.* %s .* %s\\n$",
			test.code,
			test.method,
			test.path,
		)

		matcher := regexp.MustCompile(pattern)

		doTestRequest(app, test.method, test.path)

		c.Check(matcher.Match(sw.Bytes()), Equals, true, Commentf("test %d failed regex `%s` matches `%s`", index, pattern, sw.String()))
	}
}
