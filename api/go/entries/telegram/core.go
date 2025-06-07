package handler

import (
	"net/http"

	appfx "github/huangc28/kikichoice-be/api/go/_internal/fx"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands"
	add_product "github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands/add_product"
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
			add_product.NewProductDAO,
			telegram.NewBotAPI,
		),

		// AddProductStates, should extract to a fx file
		fx.Provide(
			add_product.AsAddProductState(add_product.NewAddProductStateInit),
			add_product.AsAddProductState(add_product.NewAddProductStateSKU),
			add_product.AsAddProductState(add_product.NewAddProductStateName),
			add_product.AsAddProductState(add_product.NewAddProductStateCategory),
			add_product.AsAddProductState(add_product.NewAddProductStatePrice),
			add_product.AsAddProductState(add_product.NewAddProductStateStock),
			add_product.AsAddProductState(add_product.NewAddProductStateDescription),
			add_product.AsAddProductState(add_product.NewAddProductStateSpecs),
			add_product.AsAddProductState(add_product.NewAddProductStateImages),
			add_product.AsAddProductState(add_product.NewAddProductStateConfirm),
			add_product.AsAddProductState(add_product.NewAddProductStateCompleted),
			add_product.AsAddProductState(add_product.NewAddProductStateCancelled),
			add_product.AsAddProductState(add_product.NewAddProductStatePaused),

			fx.Annotate(
				add_product.NewAddProductStateMap,
				fx.ParamTags(`group:"add_product_states"`),
			),
		),

		fx.Provide(
			commands.AsCommandHandler(add_product.NewAddProductCommand),

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
