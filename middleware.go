package webapp

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type middlewareKey int

// ReqIDKey is the RequestID middleware key used to store the request ID value in the context.
const (
	requestIDKey middlewareKey = iota
	errorKey
	stackTraceKey
)

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

// RequestID is a middleware that injects a request ID into the context of each request.
// Retrieve it using RequestID(ctx). If the incoming request has a RequestIDHeader header then
// that value is used else a random value is generated.
func UniqueRequestID() Middleware {
	return func(next ContextHandler) ContextHandler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			id := req.Header.Get(RequestIDHeader)
			if id == "" {
				id = fmt.Sprintf("%s-%d", reqPrefix, atomic.AddInt64(&reqID, 1))
			}

			ctx = context.WithValue(ctx, requestIDKey, id)

			next(ctx, rw, req)
		}
	}
}

// RequestID retreives the request id from the context
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}

	return ""
}

// Timeout limits a request handler to run for max time of the duration
// Timeout( 15 * time.Second ) will limit the request to run no longer tan 15 seconds
// When the request times out, the request will send a 503 response
func Timeout(duration time.Duration) Middleware {
	return func(next ContextHandler) ContextHandler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {

			h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				next(ctx, rw, req)
			})

			http.TimeoutHandler(h, duration, "Server request timeout").ServeHTTP(rw, req)
		}
	}
}

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery(errorHandler ContextHandler) Middleware {
	return func(next ContextHandler) ContextHandler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stackTrace := stack(3)

					ctx = context.WithValue(ctx, errorKey, fmt.Errorf("%s", err))
					ctx = context.WithValue(ctx, stackTraceKey, stackTrace)

					if errorHandler != nil {
						errorHandler(ctx, rw, req)
					} else {
						rw.WriteHeader(http.StatusInternalServerError)
						rw.Write([]byte("Internal server error"))
					}
				}
			}()
			next(ctx, rw, req)
		}
	}
}

// ErrorStackTrace retrieves the stack trace from the context
// if a panic occurs and is handled by the recovery middleware
// a stack trace is stored in the context
//
func ErrorStackTrace(ctx context.Context) []byte {
	if ctx == nil {
		return nil
	}

	if id, ok := ctx.Value(stackTraceKey).([]byte); ok {
		return id
	}

	return nil
}

// Error retrieves the error from the context
// if a panic occurs and is handled by the recovery middleware
// the error from recovery is stored in the context
// if none is present this function will return a nil pointer
func Error(ctx context.Context) error {
	if ctx == nil {
		return nil
	}

	if id, ok := ctx.Value(errorKey).(error); ok {
		return id
	}

	return nil
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
func LogRequest(logger *log.Logger) Middleware {
	return func(next ContextHandler) ContextHandler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			// Start timer
			start := time.Now()

			//call the next handler
			next(ctx, rw, req)

			// Stop timer
			end := time.Now()
			latency := end.Sub(start)

			method := req.Method
			statusCode := rw.(ResponseWriter).Status()
			statusColor := colorForStatus(statusCode)
			methodColor := colorForMethod(method)

			logger.Printf("%v |%s %3d %s| %12v | %21s |%s %7s %s %s",
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, statusCode, reset,
				latency,
				req.RemoteAddr,
				methodColor, method, reset,
				req.URL.Path,
			)
		}
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
	default:
		return reset
	}
}
