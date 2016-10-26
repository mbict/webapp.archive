package webapp

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type App interface {
	RouteGroup

	MethodNotAllowed(handler ContextHandler)
	NotFound(handler ContextHandler)

	RedirectFixedPath(v bool)
	RedirectTrailingSlash(v bool)
	HandleOptions(v bool)

	ListenAndServe(addr string) error
	ListenAndServeTLS(addr, certFile, keyFile string) error
}

type webapp struct {
	*routeGroup
	router *httprouter.Router

	notFoundHandler   ContextHandler
	notAllowedHandler ContextHandler
}

func New() App {
	router := httprouter.New()
	group := &routeGroup{
		path:   "/",
		router: router,
	}
	router.PanicHandler = nil

	app := &webapp{
		routeGroup: group,
		router:     router,
	}

	app.HandleOptions(true)
	app.NotFound(nil)
	app.MethodNotAllowed(defaultMethodNotAllowedHandler)

	return app
}

//@todo add use function
func (app *webapp) Use(middleware ...Middleware) {
	app.routeGroup.Use(middleware...)

	//update handlers
	app.NotFound(app.notFoundHandler)
	app.MethodNotAllowed(app.notAllowedHandler)
}

func (app *webapp) MethodNotAllowed(handler ContextHandler) {
	app.notAllowedHandler = handler
	app.router.HandleMethodNotAllowed = handler != nil
	if !app.router.HandleMethodNotAllowed {
		return
	}

	methodNotAllowedHandler := app.routeGroup.middleware.Then(handler)
	app.router.MethodNotAllowed = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		methodNotAllowedHandler(context.Background(), newResponseWriter(rw), req)
	})
}

func (app *webapp) NotFound(handler ContextHandler) {
	if handler == nil {
		handler = defaultNotFoundHandler
	}
	app.notFoundHandler = handler

	notFoundHandler := app.routeGroup.middleware.Then(handler)
	app.router.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notFoundHandler(context.Background(), newResponseWriter(rw), req)
	})
}

func (app *webapp) RedirectFixedPath(v bool) {
	app.router.RedirectFixedPath = v
}

func (app *webapp) RedirectTrailingSlash(v bool) {
	app.router.RedirectTrailingSlash = v
}

func (app *webapp) HandleOptions(v bool) {
	app.router.HandleOPTIONS = v
}

// ListenAndServe starts a HTTP server and sets up a listener on the given host/port.
func (app *webapp) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, app.router)
}

// ListenAndServeTLS starts a HTTPS server and sets up a listener on the given host/port.
func (app *webapp) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, app.router)
}

func defaultMethodNotAllowedHandler(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
	http.Error(rw,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}

func defaultNotFoundHandler(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
	http.Error(rw,
		http.StatusText(http.StatusNotFound),
		http.StatusNotFound,
	)
}
