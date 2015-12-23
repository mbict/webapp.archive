package webapp

import (
	"math"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbict/render"
)

const (
	AbortIndex = math.MaxInt8 / 2
)

type (
	HandlerFunc func(*Context)

	Engine struct {
		*RouterGroup
		templateRender              render.Render
		middleware                  []HandlerFunc
		router                      *httprouter.Router
		allNotFoundHandlers         []HandlerFunc
		allMethodNotAllowedHandlers []HandlerFunc
		notFoundHandlers            []HandlerFunc
		methodNotAllowedHandlers    []HandlerFunc
	}
)

func New() *Engine {
	engine := &Engine{}
	engine.RouterGroup = &RouterGroup{
		Handlers:     nil,
		absolutePath: "/",
		engine:       engine,
	}
	engine.router = httprouter.New()

	return engine
}

func (engine *Engine) Use(middlewares ...HandlerFunc) {
	engine.RouterGroup.Use(middlewares...)
	engine.allNotFoundHandlers = engine.combineHandlers(engine.notFoundHandlers)
	engine.allMethodNotAllowedHandlers = engine.combineHandlers(engine.methodNotAllowedHandlers)
}

func (engine *Engine) MethodNotAllowed(handlers ...HandlerFunc) {
	engine.methodNotAllowedHandlers = handlers

	if len(handlers) > 0 {
		engine.router.HandleMethodNotAllowed = true
		engine.allMethodNotAllowedHandlers = engine.combineHandlers(handlers)

		engine.router.MethodNotAllowed = http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				ctx := engine.createContext(rw, req, nil, engine.allMethodNotAllowedHandlers)
				ctx.Next()
				if !ctx.Response.Written() {
					ctx.Response.WriteHeader(404)
					ctx.Response.Write([]byte("405 Method not allowed"))
				}
			})
	} else {
		engine.router.HandleMethodNotAllowed = false
		engine.allMethodNotAllowedHandlers = nil
	}
}

func (engine *Engine) NotFound(handlers ...HandlerFunc) {
	engine.notFoundHandlers = handlers
	engine.allNotFoundHandlers = engine.combineHandlers(handlers)

	engine.router.NotFound = http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			ctx := engine.createContext(rw, req, nil, engine.allNotFoundHandlers)
			ctx.Next()
			if !ctx.Response.Written() {
				ctx.Response.WriteHeader(404)
				ctx.Response.Write([]byte("404 Page not found"))
			}
		})
}

func (engine *Engine) RedirectFixedPath(v bool) {
	engine.router.RedirectFixedPath = v
}

func (engine *Engine) RedirectTrailingSlash(v bool) {
	engine.router.RedirectTrailingSlash = v
}

func (engine *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	engine.router.ServeHTTP(writer, request)
}

//Templates
func (engine *Engine) LoadDefaultTemplate(directory, layout string) {
	opt := render.TemplateOptions{
		DebugMode: IsDebugging(),
		Directory: directory,
		Layout:    layout,
	}

	engine.SetTemplateRender(render.NewTemplateRenderer(opt))
}

func (engine *Engine) SetTemplateRender(r render.Render) {
	engine.templateRender = r
}
