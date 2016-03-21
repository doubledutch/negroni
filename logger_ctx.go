package negroni

import (
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
)

// LoggerCtx is a middleware handler that logs the request as it goes in and the response as it goes out.
type LoggerCtx struct {
	// Logger inherits from log.Logger used to log messages with the Logger middleware
	*log.Logger
}

// NewLoggerCtx returns a new LoggerCtx instance
func NewLoggerCtx() *LoggerCtx {
	return &LoggerCtx{log.New(os.Stdout, "[negroni] ", 0)}
}

func (l *LoggerCtx) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
	start := time.Now()
	l.Printf("Started %s %s", r.Method, r.URL.Path)

	next(ctx, rw, r)

	res := rw.(ResponseWriter)
	l.Printf("Completed %v %s in %v", res.Status(), http.StatusText(res.Status()), time.Since(start))
}
