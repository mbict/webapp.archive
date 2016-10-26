package webapp

import (
	"context"
	"github.com/julienschmidt/httprouter"
	. "gopkg.in/check.v1"
)

type ParamsSuite struct{}

var _ = Suite(&ParamsSuite{})

func (s *ParamsSuite) TestNewContextWithParams(c *C) {
	params := httprouter.Params{}

	ctx := newContextWithParams(context.Background(), params)

	c.Assert(ctx.Value(paramsKey), DeepEquals, params)
}

func (s *ParamsSuite) TestParamsFromContext(c *C) {
	params := httprouter.Params{}
	ctx := newContextWithParams(context.Background(), params)

	paramsFromCtx := Params(ctx)

	c.Assert(paramsFromCtx, DeepEquals, params)
}

func (s *ParamsSuite) TestEmptyParamsWithNilCcontext(c *C) {
	paramsFromCtx := Params(nil)

	c.Assert(paramsFromCtx, DeepEquals, emptyParams)
}

func (s *ParamsSuite) TestEmptyParamsFromContextIfNotSetInContext(c *C) {
	paramsFromCtx := Params(context.Background())

	c.Assert(paramsFromCtx, DeepEquals, emptyParams)
}

func (s *ParamsSuite) TestEmptyParamsFromContextIfIncompatibleType(c *C) {
	wrongParams := int64(1234)
	ctx := context.WithValue(context.Background(), paramsKey, wrongParams)

	paramsFromCtx := Params(ctx)

	c.Assert(paramsFromCtx, Not(DeepEquals), wrongParams)
	c.Assert(paramsFromCtx, DeepEquals, emptyParams)
}

func (s *ParamsSuite) TestParamFromContext(c *C) {
	params := httprouter.Params{
		httprouter.Param{Key: "foo", Value: "bar"},
		httprouter.Param{Key: "biz", Value: "baz"},
	}
	ctx := newContextWithParams(context.Background(), params)

	value := Param(ctx, "biz")

	c.Assert(value, Equals, "baz")
}

func (s *ParamsSuite) TestParamFromContextEmptyIfNotFound(c *C) {
	params := httprouter.Params{}
	ctx := newContextWithParams(context.Background(), params)

	value := Param(ctx, "biz")

	c.Assert(value, Equals, "")
}

func (s *ParamsSuite) TestParamOKFromContext(c *C) {
	params := httprouter.Params{
		httprouter.Param{Key: "foo", Value: "bar"},
		httprouter.Param{Key: "biz", Value: "baz"},
	}
	ctx := newContextWithParams(context.Background(), params)

	value, ok := ParamOK(ctx, "biz")

	c.Assert(ok, Equals, true)
	c.Assert(value, Equals, "baz")
}

func (s *ParamsSuite) TestParamOKFromContextEmptyIfNotFound(c *C) {
	params := httprouter.Params{}
	ctx := newContextWithParams(context.Background(), params)

	value, ok := ParamOK(ctx, "biz")

	c.Assert(ok, Equals, false)
	c.Assert(value, Equals, "")
}
