package negroni

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
)

// CtxHandler handler is an interface that objects can implement to be registered to serve as middleware
// in the Negroni middleware stack.
// ServeHTTP should yield to the next middleware in the chain by invoking the next http.HandlerFunc
// passed in.
//
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
type CtxHandler interface {
	ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, f CtxHandlerFunc)
}

type CtxHandlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request)

func (h CtxHandlerFunc) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	h(ctx, rw, r)
}

// NextCtxHandlerFunc is an adapter to allow the use of ordinary functions as Negroni handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type NextCtxHandlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc)

func (h NextCtxHandlerFunc) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
	h(ctx, rw, r, next)
}

type ctxMiddleware struct {
	handler CtxHandler
	next    *ctxMiddleware
}

func (m ctxMiddleware) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(ctx, rw, r, m.next.ServeHTTP)
}

// WrapCtx converts a http.Handler into a negroni.CtxHandler so it can be used as a Negroni
// middleware. The next http.HandlerFunc is automatically called after the Handler
// is executed.
func WrapCtx(handler http.Handler) NextCtxHandlerFunc {
	return NextCtxHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
		handler.ServeHTTP(rw, r)
		next(ctx, rw, r)
	})
}

// CtxNegroni is a stack of Middleware Handlers that can be invoked as an http.Handler.
// Negroni middleware is evaluated in the order that they are added to the stack using
// the Use and UseHandler methods.
type CtxNegroni struct {
	middleware ctxMiddleware
	handlers   []CtxHandler
}

// NewCtx returns a new NegroniCtx instance with no middleware preconfigured.
func NewCtx(handlers ...CtxHandler) *CtxNegroni {
	return &CtxNegroni{
		handlers:   handlers,
		middleware: buildCtx(handlers),
	}
}

// ClassicCtx returns a new NegroniCtx instance with the default middleware already
// in the stack.
//
// Recovery - Panic Recovery Middleware
// Logger - Request/Response Logging
// Static - Static File Serving
func ClassicCtx() *CtxNegroni {
	return NewCtx(NewRecoveryCtx(), NewLoggerCtx(), NewStaticCtx(http.Dir("public")))
}

func (n *CtxNegroni) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	n.middleware.ServeHTTP(context.Background(), NewResponseWriter(rw), r)
}

// Use adds a Handler onto the middleware stack. Handlers are invoked in the order they are added to a Negroni.
func (n *CtxNegroni) Use(handler CtxHandler) {
	n.handlers = append(n.handlers, handler)
	n.middleware = buildCtx(n.handlers)
}

// UseFunc adds a Negroni-style handler function onto the middleware stack.
func (n *CtxNegroni) UseFunc(handlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc)) {
	n.Use(NextCtxHandlerFunc(handlerFunc))
}

// UseHandler adds a http.Handler onto the middleware stack. Handlers are invoked in the order they are added to a Negroni.
func (n *CtxNegroni) UseHandler(handler http.Handler) {
	n.Use(WrapCtx(handler))
}

// UseHandler adds a http.HandlerFunc-style handler function onto the middleware stack.
func (n *CtxNegroni) UseHandlerFunc(handlerFunc http.HandlerFunc) {
	n.UseHandler(http.HandlerFunc(handlerFunc))
}

// Run is a convenience function that runs the negroni stack as an HTTP
// server. The addr string takes the same format as http.ListenAndServe.
func (n *CtxNegroni) Run(addr string) {
	l := log.New(os.Stdout, "[negroni] ", 0)
	l.Printf("listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, n))
}

// Handlers returns a list of all the handlers in the current Negroni middleware chain.
func (n *CtxNegroni) Handlers() []CtxHandler {
	return n.handlers
}

func buildCtx(handlers []CtxHandler) ctxMiddleware {
	var next ctxMiddleware

	if len(handlers) == 0 {
		return voidCtxMiddleware()
	} else if len(handlers) > 1 {
		next = buildCtx(handlers[1:])
	} else {
		next = voidCtxMiddleware()
	}

	return ctxMiddleware{handlers[0], &next}
}

func voidCtxMiddleware() ctxMiddleware {
	return ctxMiddleware{
		NextCtxHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {}),
		&ctxMiddleware{},
	}
}
