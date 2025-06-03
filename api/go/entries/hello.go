package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	appfx "github.com/your-org/go-chi-vercel-starter/api/go/_internal/fx"
	"github.com/your-org/go-chi-vercel-starter/api/go/_internal/handlers/hello"
	"github.com/your-org/go-chi-vercel-starter/api/go/_internal/pkg/logger"
	"github.com/your-org/go-chi-vercel-starter/api/go/_internal/router"
	routerfx "github.com/your-org/go-chi-vercel-starter/api/go/_internal/router/fx"
	"go.uber.org/fx"
)

// Handle is the main entry point for the hello endpoint
// This demonstrates the basic pattern for creating Vercel serverless function handlers
func Handle(w http.ResponseWriter, r *http.Request) {
	fx.New(
		logger.TagLogger("hello"),
		appfx.CoreConfigOptions,
		routerfx.CoreRouterOptions,

		fx.Provide(
			router.AsRoute(hello.NewHelloHandler),
		),

		fx.Invoke(func(router *chi.Mux) {
			router.ServeHTTP(w, r)
		}),
	)
}
