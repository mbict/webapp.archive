package webapp

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbict/binding"
	"github.com/mbict/render"
)

type Context struct {
	request  *http.Request
	response ResponseWriter
	params   httprouter.Params
	errors   Errors

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
		response: newResponseWriter(rw),
		request:  req,
		params:   params,
		handlers: handlers,
		errors:   nil,
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

// Is aborted indicates if the aborted function is called for this context
func (ctx *Context) IsAborted() bool {
	return ctx.index == AbortIndex
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

// Request returns the http.Request
func (ctx *Context) Request() *http.Request {
	return ctx.request
}

// Method returns the request method like GET/POST etc
func (ctx *Context) Method() string {
	return ctx.request.Method
}

// Response returns the response writer it implements the http.ResponseWriter interface
func (ctx *Context) Response() ResponseWriter {
	return ctx.response
}

// Params returns the router params who are extracted from the route
// Params returns the httprouter.Params interface type
func (ctx *Context) Params() httprouter.Params {
	return ctx.params
}

// Param returns the router param by name
func (ctx *Context) Param(name string) string {
	return ctx.params.ByName(name)
}

// Header returns the response headers. It implements the http.Headers interface.
func (ctx *Context) Header() http.Header {
	return ctx.response.Header()
}

// Header returns the response headers. It implements the http.Headers interface.
func (ctx *Context) Errors() *Errors {
	return &ctx.errors
}

func (ctx *Context) ClientIP() string {
	return ctx.request.RemoteAddr
}

/************************************/
/********* PARSING REQUEST **********/
/************************************/
func (ctx *Context) Bind(obj interface{}) binding.Errors {
	return binding.Bind(obj, ctx.request)
}

func (ctx *Context) BindWith(obj interface{}, b binding.Binding) binding.Errors {
	return b.Bind(obj, ctx.request)
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

func (ctx *Context) Render(code int, render render.Render, obj ...interface{}) {
	if err := render.Render(ctx.response, code, obj...); err != nil {
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
	ctx.Render(code, ctx.engine.templateRender, append([]interface{}{name}, obj...)...)
}

// Writes the given string into the response body and sets the Content-Type to "text/plain".
func (ctx *Context) String(code int, format string, values ...interface{}) {
	ctx.Render(code, render.Plain, append([]interface{}{format}, values...)...)
}

// Writes the given string into the response body and sets the Content-Type to "text/html" without template.
func (ctx *Context) HTMLString(code int, format string, values ...interface{}) {
	ctx.Render(code, render.HtmlPlain, append([]interface{}{format}, values...)...)
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
		ctx.response.Header().Set("Content-Type", contentType)
	}
	ctx.WriteHeader(code)
	ctx.Write(data)
}

// Respond writes the given HTTP status code and response body.
// This method should only be called once per request.
func (ctx *Context) Respond(code int, body []byte) error {
	ctx.WriteHeader(code)
	if _, err := ctx.Write(body); err != nil {
		return err
	}
	return nil
}

// Writes the specified file into the body stream
func (ctx *Context) File(filepath string) {
	http.ServeFile(ctx.response, ctx.request, filepath)
}

// ResponseWritten returns if the response have been written to the output
func (ctx *Context) ResponseWritten() bool {
	return ctx.response.Written()
}

// ResponseStatus shows the status who is beeing send out
func (ctx *Context) ResponseStatus() int {
	return ctx.response.Status()
}

// ResponseLength returns the response body length in bytes if the response was written
func (ctx *Context) ResponseLength() int {
	return ctx.response.Size()
}

// WriteHeader writes the HTTP status code to the response. It implements the
// http.ResponseWriter interface.
func (ctx *Context) WriteHeader(code int) {
	ctx.response.WriteHeader(code)
}

// Write writes the HTTP response body. It implements the http.ResponseWriter interface.
func (ctx *Context) Write(body []byte) (int, error) {
	return ctx.response.Write(body)
}
