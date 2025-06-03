package clerk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/configs"
)

type ClerkClient struct {
	clerkCfg *clerk.ClientConfig
}

func NewClerkClient(cfg *configs.Config) *ClerkClient {
	config := &clerk.ClientConfig{}
	config.Key = &cfg.Clerk.SecretKey
	clerk.SetKey(cfg.Clerk.SecretKey)

	return &ClerkClient{
		clerkCfg: config,
	}
}

type VerifyTokenResponse struct {
	ID string
}

// VerifyToken verifies a JWT token and returns the user ID
func (c *ClerkClient) VerifyToken(ctx context.Context, token string) (VerifyTokenResponse, error) {
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: token,
	})
	if err != nil {
		return VerifyTokenResponse{}, c.formatClerkError(err, "failed to verify token")
	}

	userClient := user.NewClient(c.clerkCfg)
	usr, err := userClient.Get(ctx, claims.Subject)
	if err != nil {
		return VerifyTokenResponse{}, c.formatClerkError(err, "failed to get user")
	}

	if usr == nil {
		return VerifyTokenResponse{}, fmt.Errorf("user not found")
	}

	return VerifyTokenResponse{
		ID: usr.ID,
	}, nil
}

//	{
//	  "errors": [
//	    {
//	      "code": "already_a_member_in_organization",
//	      "message": "already a member",
//	      "long_message": "User is already a member of the organization."
//	    }
//	  ],
//	  "status": 400,
//	  "clerk_trace_id": "e180895c3295dc98ec2efe271a5185b2"
//	}
type ClerkError struct {
	Errors []struct {
		Code        string `json:"code"`
		Message     string `json:"message"`
		LongMessage string `json:"long_message"`
	} `json:"errors"`
	Status       int    `json:"status"`
	ClerkTraceID string `json:"clerk_trace_id"`
}

// formatClerkError attempts to extract structured error information from Clerk API errors.
// It unmarshals JSON error responses into a more detailed format when possible.
func (c *ClerkClient) formatClerkError(err error, customMessage string) error {
	// Try to unmarshal detailed Clerk error
	var clerkErr ClerkError
	errorStr := err.Error()
	if unmarshalErr := json.Unmarshal([]byte(errorStr), &clerkErr); unmarshalErr == nil && len(clerkErr.Errors) > 0 {
		errDetail := clerkErr.Errors[0]
		return fmt.Errorf("clerk error: code=%s, message=%s, trace_id=%s: %w",
			errDetail.Code, errDetail.LongMessage, clerkErr.ClerkTraceID, err)
	}
	// Fallback to original error if unmarshaling fails
	return fmt.Errorf("%s: %w", customMessage, err)
}
