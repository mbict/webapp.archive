package webapp

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbict/binding"
	"github.com/mbict/render"
)

type Context struct {
	Request  *http.Request
	Response ResponseWriter
	Params   httprouter.Params
	Errors   Errors

	values   map[interface{}]interface{}
	handlers []HandlerFunc
	index    int8
	engine   *Engine
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/

func (engine *Engine) createContext(rw http.ResponseWriter, req *http.Request, params httprouter.Params, handlers []HandlerFunc) *Context {
	return &Context{
		Response: newResponseWriter(rw),
		Request:  req,
		Params:   params,
		handlers: handlers,
		Errors:   nil,
		index:    -1,
		engine:   engine,
	}
}

/************************************/
/*************** FLOW ***************/
/************************************/

// Next should be used only in the middlewares.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
func (ctx *Context) Next() {
	ctx.index++
	s := int8(len(ctx.handlers))
	for ; ctx.index < s; ctx.index++ {
		ctx.handlers[ctx.index](ctx)
	}
}

// Forces the system to do not continue calling the pending handlers in the chain.
func (ctx *Context) Abort() {
	ctx.index = AbortIndex
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Sets a new pair key/value just for the specified context.
// It also lazy initializes the hashmap.
func (ctx *Context) Set(key interface{}, item interface{}) {
	if ctx.values == nil {
		ctx.values = make(map[interface{}]interface{})
	}
	ctx.values[key] = item
}

// Get returns the value for the given key or an error if the key does not exist.
func (ctx *Context) Get(key interface{}) interface{} {
	if ctx.values != nil {
		value, ok := ctx.values[key]
		if ok {
			return value
		}
	}
	return nil
}

func (ctx *Context) GetOk(key interface{}) (interface{}, bool) {
	if ctx.values != nil {
		value, ok := ctx.values[key]
		if ok {
			return value, true
		}
	}
	return nil, false
}

func (ctx *Context) ClientIP() string {
	return ctx.Request.RemoteAddr
}

/************************************/
/********* PARSING REQUEST **********/
/************************************/
func (ctx *Context) Bind(obj interface{}) binding.Errors {
	return binding.Bind(obj, ctx.Request)
}

func (ctx *Context) BindWith(obj interface{}, b binding.Binding) binding.Errors {
	return b.Bind(obj, ctx.Request)
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

func (ctx *Context) Render(code int, render render.Render, obj ...interface{}) {
	if err := render.Render(ctx.Response, code, obj...); err != nil {
		panic(err)
	}
}

// Serializes the given struct as JSON into the response body in a fast and efficient way.
// It also sets the Content-Type as "application/json".
func (ctx *Context) JSON(code int, obj interface{}) {
	ctx.Render(code, render.JSON, obj)
}

// Serializes the given struct as XML into the response body in a fast and efficient way.
// It also sets the Content-Type as "application/xml".
func (ctx *Context) XML(code int, obj interface{}) {
	ctx.Render(code, render.XML, obj)
}

// Renders the HTTP template specified by its file name.
// It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (ctx *Context) HTML(code int, name string, obj ...interface{}) {
	ctx.Render(code, ctx.engine.templateRender, name, obj)
}

// Writes the given string into the response body and sets the Content-Type to "text/plain".
func (ctx *Context) String(code int, format string, values ...interface{}) {
	ctx.Render(code, render.Plain, format, values)
}

// Writes the given string into the response body and sets the Content-Type to "text/html" without template.
func (ctx *Context) HTMLString(code int, format string, values ...interface{}) {
	ctx.Render(code, render.HtmlPlain, format, values)
}

// Returns a HTTP redirect to the specific location.
func (ctx *Context) Redirect(code int, location string) {
	if code >= 300 && code <= 308 {
		ctx.Render(code, render.Redirect, location)
	} else {
		panic(fmt.Sprintf("Cannot send a redirect with status code %d", code))
	}
}

// Writes some data into the body stream and updates the HTTP code.
func (ctx *Context) Data(code int, contentType string, data []byte) {
	if len(contentType) > 0 {
		ctx.Response.Header().Set("Content-Type", contentType)
	}
	ctx.Response.WriteHeader(code)
	ctx.Response.Write(data)
}

// Writes the specified file into the body stream
func (ctx *Context) File(filepath string) {
	http.ServeFile(ctx.Response, ctx.Request, filepath)
}
