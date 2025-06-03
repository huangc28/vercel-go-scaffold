package pkgfx

import (
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/configs"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/db"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/middlewares"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/pkg/clerk"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/pkg/logger"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/pkg/trino"
	"go.uber.org/fx"
)

var AuthMiddlewareModule = fx.Module(
	"auth-middleware",
	fx.Provide(
		fx.Private,
		configs.NewViper,
		configs.NewConfig,
		clerk.NewClerkClient,
	),
	fx.Provide(
		middlewares.NewAuthMiddleware,
	),
)

var TrinoModule = fx.Module(
	"trino-client",
	fx.Provide(
		fx.Private,
		configs.NewViper,
		db.NewTrinoDB,
		configs.NewConfig,
		logger.NewLogger,
	),
	fx.Provide(
		trino.NewTrinoClient,
	),
)
