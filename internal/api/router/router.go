package router

import (
	"net/http"

	"github.com/rizkicandra/dandanna-api/internal/api/middleware"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// HealthHandler is the interface the router depends on — not the concrete type.
type HealthHandler interface {
	Handle(http.ResponseWriter, *http.Request)
	Readyz(http.ResponseWriter, *http.Request)
}

// Router manages HTTP routing
type Router struct {
	mux         *http.ServeMux
	logger      logger.Logger
	corsOrigins []string
}

// New creates a new router instance
func New(log logger.Logger, corsOrigins []string) *Router {
	return &Router{
		mux:         http.NewServeMux(),
		logger:      log,
		corsOrigins: corsOrigins,
	}
}

// Setup registers all routes
func (r *Router) Setup(health HealthHandler) {
	r.mux.HandleFunc("GET /api/health", health.Handle)
	r.mux.HandleFunc("GET /api/readyz", health.Readyz)

	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.NotFound(w, req)
	})
}

// Handler returns the fully middleware-wrapped HTTP handler.
// Middlewares are applied top-to-bottom: Recovery is outermost, Logging is innermost.
func (r *Router) Handler() http.Handler {
	return chain(r.mux,
		middleware.Recovery(r.logger),
		middleware.SecurityHeaders,
		middleware.CORS(r.corsOrigins),
		middleware.RequestID,
		middleware.Logging(r.logger),
	)
}

// chain applies middlewares to h in order: the first middleware is the outermost wrapper.
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
