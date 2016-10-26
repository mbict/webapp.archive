package webapp

import (
	"context"
	"github.com/julienschmidt/httprouter"
)

type key int

const paramsKey key = iota

var emptyParams = httprouter.Params{}

func newContextWithParams(ctx context.Context, params httprouter.Params) context.Context {
	return context.WithValue(ctx, paramsKey, params)
}

// Params returns URL parameters stored in context
func Params(ctx context.Context) httprouter.Params {
	if ctx == nil {
		return emptyParams
	}
	if p, ok := ctx.Value(paramsKey).(httprouter.Params); ok {
		return p
	}
	return emptyParams
}

// Param picks one URL parameters stored in context by its name.
//
// This is a shortcut for:
//   xmux.Params(ctx).ByName("name")
func Param(ctx context.Context, name string) string {
	return Params(ctx).ByName(name)
}

// ParamOK picks one URL parameters stored in context by its name.
func ParamOK(ctx context.Context, name string) (string, bool) {
	params := Params(ctx)
	for i := range params {
		if params[i].Key == name {
			return params[i].Value, true
		}
	}
	return "", false
}
