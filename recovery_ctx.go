package negroni

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"golang.org/x/net/context"
)

// RecoveryCtx is a Negroni middleware that recovers from any panics and writes a 500 if there was one.
type RecoveryCtx struct {
	Logger     *log.Logger
	PrintStack bool
	StackAll   bool
	StackSize  int
}

// NewRecoveryCtx returns a new instance of RecoveryCtx
func NewRecoveryCtx() *RecoveryCtx {
	return &RecoveryCtx{
		Logger:     log.New(os.Stdout, "[negroni] ", 0),
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

func (rec *RecoveryCtx) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]

			f := "PANIC: %s\n%s"
			rec.Logger.Printf(f, err, stack)

			if rec.PrintStack {
				fmt.Fprintf(rw, f, err, stack)
			}
		}
	}()

	next(ctx, rw, r)
}
