package webapp

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type middlewareKey int

// ReqIDKey is the RequestID middleware key used to store the request ID value in the context.
const RequestIDKey middlewareKey = 0

// RequestIDHeader is the name of the header used to transmit the request ID.
const RequestIDHeader = "X-Request-Id"

// Counter used to create new request ids.
var reqID int64

// Common prefix to all newly created request ids for this process.
var reqPrefix string

// Initialize common prefix on process startup.
func init() {
	// algorithm taken from https://github.com/zenazn/goji/blob/master/web/middleware/request_id.go#L44-L50
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}
	reqPrefix = string(b64[0:10])
}

//func NewMiddleware(m interface{}) (mw Middleware, err error) {
//	switch m := m.(type) {
//	case Middleware:
//		mw = m
//	case func(Handler) Handler:
//		mw = m
//	case Handler:
//		mw = handlerToMiddleware(m)
//	case func(*Context) error:
//		mw = handlerToMiddleware(m)
//	case func(http.Handler) http.Handler:
//		mw = func(h Handler) Handler {
//			return func(ctx *Context) (err error) {
//				rw := ctx.Value(respKey).(http.ResponseWriter)
//				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//					err = h(ctx)
//				})).ServeHTTP(rw, ctx.Request())
//				return
//			}
//		}
//	case http.Handler:
//		mw = httpHandlerToMiddleware(m.ServeHTTP)
//	case func(http.ResponseWriter, *http.Request):
//		mw = httpHandlerToMiddleware(m)
//	default:
//		err = fmt.Errorf("invalid middleware %#v", m)
//	}
//	return
//}

// RequestID is a middleware that injects a request ID into the context of each request.
// Retrieve it using ctx.Value(ReqIDKey). If the incoming request has a RequestIDHeader header then
// that value is used else a random value is generated.
func RequestID() HandlerFunc {
	return func(ctx *Context) {
		id := ctx.Header().Get(RequestIDHeader)
		if id == "" {
			id = fmt.Sprintf("%s-%d", reqPrefix, atomic.AddInt64(&reqID, 1))
		}
		ctx.Set(RequestIDKey, id)
	}
}

// Timeout limits a request handler to run for max time of the duration
// Timeout( 15 * time.Second ) will limit the request to run no longer tan 15 seconds
// When the request times out, the request will send a 503 response
func Timeout(duration time.Duration) HandlerFunc {
	return func(ctx *Context) {
		h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx.Next()
		})

		http.TimeoutHandler(h, duration, "Server request timeout").ServeHTTP(ctx.Response(), ctx.Request())
	}
}

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery(handler HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				//abort context
				ctx.Abort()

				stackTrace := stack(3)
				formattedError := fmt.Errorf("%s\nStack trace:\n%s", err, stackTrace)
				ctx.Errors().Add(formattedError)

				if handler != nil {
					handler(ctx)
				}
				if !ctx.ResponseWritten() {
					ctx.Respond(http.StatusInternalServerError, []byte("Internal server error"))
				}
			}
		}()
		ctx.Next()
	}
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// stack returns a nicely formated stack frame, skipping skip frames
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

// LogRequest creates a request logger middleware.
//
func LogRequest(out io.Writer) HandlerFunc {
	return func(ctx *Context) {
		/*
			return func(ctx *Context) error {
				reqID := ctx.Value(ReqIDKey)
				if reqID == nil {
					reqID = shortID()
				}
				ctx.Logger = ctx.Logger.New("id", reqID)
				startedAt := time.Now()
				r := ctx.Value(reqKey).(*http.Request)
				ctx.Info("started", r.Method, r.URL.String())
				params := ctx.Value(paramKey).(map[string]string)
				if len(params) > 0 {
					logCtx := make(log.Ctx, len(params))
					for k, v := range params {
						logCtx[k] = interface{}(v)
					}
					ctx.Debug("params", logCtx)
				}
				query := ctx.Value(queryKey).(map[string][]string)
				if len(query) > 0 {
					logCtx := make(log.Ctx, len(query))
					for k, v := range query {
						logCtx[k] = interface{}(v)
					}
					ctx.Debug("query", logCtx)
				}
				payload := ctx.Value(payloadKey)
				if r.ContentLength > 0 {
					if mp, ok := payload.(map[string]interface{}); ok {
						ctx.Debug("payload", log.Ctx(mp))
					} else {
						ctx.Debug("payload", "raw", payload)
					}
				}
				err := h(ctx)
				ctx.Info("completed", "status", ctx.ResponseStatus(),
					"bytes", ctx.ResponseLength(), "time", time.Since(startedAt).String())
				return err
			}
		*/

		// Start timer
		start := time.Now()

		ctx.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		method := ctx.Method()
		statusCode := ctx.ResponseStatus()
		statusColor := colorForStatus(statusCode)
		methodColor := colorForMethod(method)

		fmt.Fprintf(out, "[LOG] %v |%s %3d %s| %12v | %21s |%s %7s %s %s\n%s",
			end.Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, reset,
			latency,
			ctx.ClientIP(),
			methodColor, method, reset,
			ctx.Request().URL.Path,
			ctx.Errors().String(),
		)
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code <= 299:
		return green
	case code >= 300 && code <= 399:
		return white
	case code >= 400 && code <= 499:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch {
	case method == "GET":
		return blue
	case method == "POST":
		return cyan
	case method == "PUT":
		return yellow
	case method == "DELETE":
		return red
	case method == "PATCH":
		return green
	case method == "HEAD":
		return magenta
	case method == "OPTIONS":
		return white
	case method == "CONNECT":
		return white
	default:
		return reset
	}
}
