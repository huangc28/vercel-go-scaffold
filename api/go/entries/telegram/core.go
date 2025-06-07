package handler

import (
	"net/http"

	appfx "github/huangc28/kikichoice-be/api/go/_internal/fx"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands"
	"github/huangc28/kikichoice-be/api/go/_internal/pkg/logger"
	"github/huangc28/kikichoice-be/api/go/_internal/router"
	routerfx "github/huangc28/kikichoice-be/api/go/_internal/router/fx"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	fx.New(
		logger.TagLogger("telegram"),
		appfx.CoreConfigOptions,
		routerfx.CoreRouterOptions,

		fx.Provide(
			commands.NewCommandDAO,
			commands.NewProductDAO,
			telegram.NewBotAPI,
		),

		fx.Provide(
			commands.AsCommandHandler(commands.NewAddProductCommand),

			fx.Annotate(
				commands.NewCommandHandlerMap,
				fx.ParamTags(`group:"command_handlers"`),
			),
		),

		fx.Provide(
			router.AsRoute(telegram.NewTelegramHandler),
		),

		fx.Invoke(func(router *chi.Mux) {
			router.ServeHTTP(w, r)
		}),
	)
}
