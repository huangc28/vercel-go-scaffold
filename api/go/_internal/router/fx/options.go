package routerfx

import (
	"github/huangc28/kikichoice-be/api/go/_internal/router"
	"go.uber.org/fx"
)

var CoreRouterOptions = fx.Options(
	fx.Provide(
		fx.Annotate(
			router.NewRouter,
			fx.ParamTags(`group:"handlers"`),
		),
	),
)
