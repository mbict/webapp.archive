package webapp

import (
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"path"
	"reflect"
	"runtime"
)

type (
	RouteGroup interface {
		Logger(logger *log.Logger)

		Use(middleware ...Middleware)
		With(middleware ...Middleware) RouteGroup
		Group(relativePath string, middleware ...Middleware) RouteGroup

		POST(relativePath string, handler ContextHandler)
		GET(relativePath string, handler ContextHandler)
		DELETE(relativePath string, handler ContextHandler)
		PATCH(relativePath string, handler ContextHandler)
		PUT(relativePath string, handler ContextHandler)
		OPTIONS(relativePath string, handler ContextHandler)
		HEAD(relativePath string, handler ContextHandler)
		LINK(relativePath string, handler ContextHandler)
		UNLINK(relativePath string, handler ContextHandler)

		Static(relativePath, directory string)
		StaticFile(relativePath, file string)

		Handle(httpMethod, relativePath string, handler ContextHandler)
		ServeHTTP(rw http.ResponseWriter, req *http.Request)
	}

	routeGroup struct {
		path       string
		middleware Chain
		router     *httprouter.Router
		logger     *log.Logger
	}
)

// NewRouteGroup creates a new RouteGroup handler
func newRouteGroup(router *httprouter.Router) RouteGroup {
	return &routeGroup{
		path:   "/",
		router: router,
	}
}

// Logger will set the logger to report to when debugging the created routes
func (group *routeGroup) Logger(logger *log.Logger) {
	group.logger = logger
}

// Use pushes middleware on the middleware chain
// be aware that already added routes are not updated
func (group *routeGroup) Use(middleware ...Middleware) {
	group.middleware = group.middleware.Append(middleware...)
}

// With creates a new Group with the same path and pushes the new middleware to stack
// without modify the existing  group
func (group *routeGroup) With(middleware ...Middleware) RouteGroup {
	return &routeGroup{
		path:       group.path,
		middleware: group.middleware.Append(middleware...),
		router:     group.router,
		logger:     group.logger,
	}
}

// Group creates a new sub route and copies
func (group *routeGroup) Group(relativePath string, middleware ...Middleware) RouteGroup {
	return &routeGroup{
		path:       group.calculateAbsolutePath(relativePath),
		middleware: group.middleware.Append(middleware...),
		router:     group.router,
		logger:     group.logger,
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (group *routeGroup) GET(relativePath string, handler ContextHandler) {
	group.Handle("GET", relativePath, handler)
}

// POST is a shortcut for router.Handle("POST", relativePath, handle)
func (group *routeGroup) POST(relativePath string, handler ContextHandler) {
	group.Handle("POST", relativePath, handler)
}

// DELETE is a shortcut for router.Handle("DELETE", relativePath, handle)
func (group *routeGroup) DELETE(relativePath string, handler ContextHandler) {
	group.Handle("DELETE", relativePath, handler)
}

// PATCH is a shortcut for router.Handle("PATCH", relativePath, handle)
func (group *routeGroup) PATCH(relativePath string, handler ContextHandler) {
	group.Handle("PATCH", relativePath, handler)
}

// PUT is a shortcut for router.Handle("PUT", relativePath, handle)
func (group *routeGroup) PUT(relativePath string, handler ContextHandler) {
	group.Handle("PUT", relativePath, handler)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", relativePath, handle)
func (group *routeGroup) OPTIONS(relativePath string, handler ContextHandler) {
	group.Handle("OPTIONS", relativePath, handler)
}

// HEAD is a shortcut for router.Handle("HEAD", relativePath, handle)
func (group *routeGroup) HEAD(relativePath string, handler ContextHandler) {
	group.Handle("HEAD", relativePath, handler)
}

// LINK is a shortcut for router.Handle("LINK", relativePath, handle)
func (group *routeGroup) LINK(relativePath string, handler ContextHandler) {
	group.Handle("LINK", relativePath, handler)
}

// UNLINK is a shortcut for router.Handle("UNLINK", relativePath, handle)
func (group *routeGroup) UNLINK(relativePath string, handler ContextHandler) {
	group.Handle("UNLINK", relativePath, handler)
}

// Handle registers a new request handle and middlewares with the given path and method.
// The last handler should be the real handler, the other ones should be middlewares that can and should be shared among different routes.
// See the example code in github.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.group. for internal
// communication with a proxy).
func (group *routeGroup) Handle(httpMethod, relativePath string, handler ContextHandler) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handler = group.middleware.Then(handler)

	//debug route logging
	if group.logger != nil {
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		group.logger.Printf("%-7s %-35s --> %s (%d middlewares)\n", httpMethod, absolutePath, handlerName, len(group.middleware))
	}

	group.router.Handle(httpMethod, absolutePath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := newContextWithParams(context.Background(), params)
		rw = newResponseWriter(rw)
		handler(ctx, rw, req)
	})
}

// ServeHTTP
func (group *routeGroup) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	group.router.ServeHTTP(rw, req)
}

// Static serves files from the given file system root.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use :
//     router.Static("/static", "/var/www")
func (group *routeGroup) Static(relativePath, root string) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handler := createStaticHandler(absolutePath, root)
	relativePath = path.Join(relativePath, "/*filepath")

	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
}

func createStaticHandler(absolutePath, root string) ContextHandler {
	fileServer := http.StripPrefix(absolutePath, http.FileServer(http.Dir(root)))
	return func(_ context.Context, rw http.ResponseWriter, req *http.Request) {
		fileServer.ServeHTTP(rw, req)
	}
}

func (group *routeGroup) StaticFile(relativePath, file string) {
	handler := createStaticFileHandler(file)

	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
}

func createStaticFileHandler(file string) ContextHandler {
	return func(_ context.Context, rw http.ResponseWriter, req *http.Request) {
		http.ServeFile(rw, req, file)
	}
}

func (group *routeGroup) calculateAbsolutePath(relativePath string) string {
	if len(relativePath) == 0 {
		return group.path
	}

	absolutePath := path.Join(group.path, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(absolutePath) != '/'
	if appendSlash {
		return absolutePath + "/"
	}
	return absolutePath
}

func lastChar(str string) uint8 {
	size := len(str)
	return str[size-1]
}
