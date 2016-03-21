package negroni

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

func TestNegroniCtxRun(t *testing.T) {
	// just test that Run doesn't bomb
	go NewCtx().Run(":3001")
}

func TestNegroniCtxServeHTTP(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	n := NewCtx()
	n.Use(NextCtxHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
		result += "foo"
		next(ctx, rw, r)
		result += "ban"
	}))
	n.Use(NextCtxHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
		result += "bar"
		next(ctx, rw, r)
		result += "baz"
	}))
	n.Use(NextCtxHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
		result += "bat"
		rw.WriteHeader(http.StatusBadRequest)
	}))

	n.ServeHTTP(response, (*http.Request)(nil))

	expect(t, result, "foobarbatbazban")
	expect(t, response.Code, http.StatusBadRequest)
}

// Ensures that a NegroniCtx middleware chain
// can correctly return all of its handlers.
func TestNegroniCtxHandlers(t *testing.T) {
	response := httptest.NewRecorder()
	n := NewCtx()
	handlers := n.Handlers()
	expect(t, 0, len(handlers))

	n.Use(NextCtxHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next CtxHandlerFunc) {
		rw.WriteHeader(http.StatusOK)
	}))

	// Expects the length of handlers to be exactly 1
	// after adding exactly one handler to the middleware chain
	handlers = n.Handlers()
	expect(t, 1, len(handlers))

	// Ensures that the first handler that is in sequence behaves
	// exactly the same as the one that was registered earlier
	handlers[0].ServeHTTP(context.Background(), response, (*http.Request)(nil), nil)
	expect(t, response.Code, http.StatusOK)
}
