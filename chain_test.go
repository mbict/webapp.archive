package webapp

import (
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
)

type ChainSuite struct{}

var _ = Suite(&ChainSuite{})

func (s *ChainSuite) TestNewChainLength(c *C) {
	chain := NewChain(middlewareWriter("1"), middlewareWriter("2"))

	c.Assert(chain, HasLen, 2)
}

func (s *ChainSuite) TestNewEmptyChainLength(c *C) {
	chain := NewChain()

	c.Assert(chain, HasLen, 0)
}

func (s *ChainSuite) TestChainExecutionOrderFIFO(c *C) {
	rw := httptest.NewRecorder()
	chain := NewChain(middlewareWriter("1"), middlewareWriter("2"))
	h := chain.Then(finalHandler)

	h(context.Background(), rw, (*http.Request)(nil))

	c.Assert(rw.Body.String(), Equals, "12H")
}

func (s *ChainSuite) TestEmptyChain(c *C) {
	rw := httptest.NewRecorder()
	chain := NewChain()
	h := chain.Then(finalHandler)

	h(context.Background(), rw, (*http.Request)(nil))

	c.Assert(rw.Body.String(), Equals, "H")
}

func (s *ChainSuite) TestMiddlewareAborts(c *C) {
	rw := httptest.NewRecorder()
	chain := NewChain(middlewareWriter("1"), middlewareAbort())
	h := chain.Then(finalHandler)

	h(context.Background(), rw, (*http.Request)(nil))

	c.Assert(rw.Body.String(), Equals, "1abort")
}

func (s *ChainSuite) TestAppend(c *C) {
	chain1 := NewChain(middlewareWriter("1"))
	chain2 := chain1.Append(middlewareWriter("2"))

	rw := httptest.NewRecorder()
	chain2.Then(finalHandler)(context.Background(), rw, (*http.Request)(nil))
	c.Assert(rw.Body.String(), Equals, "12H")
}

func (s *ChainSuite) TestAppendRespectsImmutableChain(c *C) {
	chain1 := NewChain(middlewareWriter("1"))
	chain2 := chain1.Append(middlewareWriter("1"))

	c.Assert(chain1, HasLen, 1)
	c.Assert(chain2, HasLen, 2)

	c.Assert(chain1, Not(DeepEquals), chain2)
}

func (s *ChainSuite) TestExtend(c *C) {
	chain1 := NewChain(middlewareWriter("1"), middlewareWriter("2"))
	chain2 := NewChain(middlewareWriter("3"), middlewareWriter("4"))
	chainExtend := chain1.Extend(chain2)

	rw := httptest.NewRecorder()
	chainExtend.Then(finalHandler)(context.Background(), rw, (*http.Request)(nil))

	c.Assert(rw.Body.String(), Equals, "1234H")
}

func (s *ChainSuite) TestExtendRespectsImmutableChain(c *C) {
	chain1 := NewChain(middlewareWriter("1"), middlewareWriter("2"))
	chain2 := NewChain(middlewareWriter("3"), middlewareWriter("4"))
	chainExtend := chain1.Extend(chain2)

	c.Assert(chain1, HasLen, 2)
	c.Assert(chain2, HasLen, 2)
	c.Assert(chainExtend, HasLen, 4)

	c.Assert(chainExtend, Not(DeepEquals), chain1)
	c.Assert(chainExtend, Not(DeepEquals), chain2)
}

func (s *ChainSuite) TestWrapMiddlewareHTTP(c *C) {
	tests := []struct {
		handler     interface{}
		description string
		result      string
	}{
		{
			handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Write([]byte("M"))
			}),
			description: "http.Handler",
			result:      "1M2H",
		}, {
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.Write([]byte("M"))
			},
			description: "http.HandlerFunc",
			result:      "1M2H",
		}, {
			handler: func(http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.Write([]byte("M"))
				})
			},
			description: "httpMiddleware : func(http.Handler) http.Handler",
			result:      "1M2H",
		}, {
			handler: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
				rw.Write([]byte("M"))
			},
			description: "ContextHandlerFunc",
			result:      "1M2H",
		}, {
			handler: func(ContextHandler) ContextHandler {
				return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
					rw.Write([]byte("M"))
				}
			},
			description: "Middleware / (stops execution of handler on the middleware)",
			result:      "1M",
		}, {
			handler: func(next ContextHandler) ContextHandler {
				return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
					rw.Write([]byte("M"))
					next(ctx, rw, req)
				}
			},
			description: "Middleware",
			result:      "1M2H",
		},
	}

	for _, t := range tests {

		m := (Middleware)(nil)
		if t.handler != nil {
			m = WrapMiddleware(t.handler)
		}

		rw := httptest.NewRecorder()
		h := NewChain(middlewareWriter("1"), m, middlewareWriter("2")).Then(finalHandler)
		h(nil, rw, (*http.Request)(nil))

		c.Check(rw.Body.String(), Equals, t.result, Commentf("wrapping of `%s` failed", t.description))
	}
}

func (s *ChainSuite) TestWrapMiddlewareUnkownType(c *C) {
	wrapPanic := func() {
		WrapMiddleware(func(a int) {})
	}
	c.Assert(wrapPanic, PanicMatches, `unsupported handler\: func\(int\)`)
}
