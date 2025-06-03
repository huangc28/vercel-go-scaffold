package routerfx

import (
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/router"
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
