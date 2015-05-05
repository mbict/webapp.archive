package webapp

import (
	"net/http"
	"reflect"
	"runtime"
)

func lastChar(str string) uint8 {
	size := len(str)
	if size == 0 {
		panic("The length of the string can't be 0")
	}
	return str[size-1]
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func HttpHandlerFunc(h http.HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		h(ctx.Response, ctx.Request)
	}
}
