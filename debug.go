package webapp

import (
	"log"
	"os"
)

var debugLogger = log.New(os.Stdout, "[WEBAPP-debug] ", 0)

func IsDebugging() bool {
	return env == debugCode
}

func debugRoute(httpMethod, absolutePath string, handlers []HandlerFunc) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers[nuHandlers-1])
		debugPrint("%-7s %-35s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, len(handlers))
	}
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		debugLogger.Printf(format, values...)
	}
}
