package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/pkg/clerk"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/pkg/render"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// extractBearerToken helper function remains the same
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be 'Bearer {token}'")
	}

	return parts[1], nil
}

type AuthMiddleware struct {
	clerkcli *clerk.ClerkClient
	logger   *zap.SugaredLogger
}

type AuthMiddlewareParams struct {
	fx.In

	ClerkClient *clerk.ClerkClient
	Logger      *zap.SugaredLogger
}

func NewAuthMiddleware(p AuthMiddlewareParams) *AuthMiddleware {
	return &AuthMiddleware{
		clerkcli: p.ClerkClient,
		logger:   p.Logger,
	}
}

func (m *AuthMiddleware) Authed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Error("Missing Authorization header")
			render.ChiErr(
				w, r,
				fmt.Errorf("missing Authorization header"),
				MissingAuthorizationHeader,
				render.WithStatusCode(http.StatusUnauthorized),
			)
			return
		}

		// Extract token
		token, err := extractBearerToken(authHeader)
		if err != nil {
			// m.logger.Errorw("Failed to extract bearer token", "error", err)
			render.ChiErr(
				w, r,
				err,
				FailedToExtractBearerToken,
				render.WithStatusCode(http.StatusUnauthorized),
			)
			return
		}

		// Verify token with Clerk
		verifyTokenResp, err := m.clerkcli.VerifyToken(r.Context(), token)
		if err != nil {
			m.logger.Errorw(
				"Failed to verify token",
				"error", err,
			)
			render.ChiErr(
				w, r,
				err,
				InvalidBearerToken,
				render.WithStatusCode(http.StatusUnauthorized),
			)
			return
		}

		m.logger.Infof("Auth provider ID %s", verifyTokenResp.ID)

		next.ServeHTTP(w, r)
	})
}
