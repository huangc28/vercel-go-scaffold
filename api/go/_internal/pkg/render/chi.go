package render

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

type ErrReason struct {
	Err error `json:"-"` // low-level runtime error

	StatusText string `json:"status_text"`     // user-level status message
	AppCode    string `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e ErrReason) Error() string {
	return fmt.Sprintf("app_code: %s, status_text: %s, error: %s", e.AppCode, e.StatusText, e.ErrorText)
}

func (e ErrReason) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// ChiResponse mirrors ApiGatewayResponse but for Chi renderer
type ChiResponse struct {
	HTTPStatusCode int `json:"-"`

	Data   any         `json:"data"`
	Errors []ErrReason `json:"errors"`
}

// Render implements the render.Renderer interface
func (resp ChiResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, resp.HTTPStatusCode)
	return nil
}

func WithStatusCode(statusCode int) func(opt *ChiResponse) {
	return func(opt *ChiResponse) {
		opt.HTTPStatusCode = statusCode
	}
}

// ChiJSON is the Chi equivalent of ApiGatewayJSON
// It renders a JSON response using Chi render with the same structure as ApiGatewayJSON
func ChiJSON(w http.ResponseWriter, r *http.Request, data any, opts ...func(opt *ChiResponse)) {
	// Create the response with the same structure
	resp := ChiResponse{
		HTTPStatusCode: http.StatusOK,
		Data:           data,
		Errors:         nil,
	}

	// Apply options just like in ApiGatewayJSON
	for _, opt := range opts {
		opt(&resp)
	}

	// Set the status code
	render.Status(r, resp.HTTPStatusCode)

	// Render the response
	render.JSON(w, r, resp)
}

// ChiErr is the Chi equivalent of APIGatewayErr
func ChiErr(w http.ResponseWriter, r *http.Request, err error, appCode string, opts ...func(er *ChiResponse)) {
	errResp := ErrReason{
		Err:        err,
		StatusText: "",
		AppCode:    appCode,
		ErrorText:  err.Error(),
	}

	resp := ChiResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		Data:           nil,
		Errors:         []ErrReason{errResp},
	}

	for _, opt := range opts {
		opt(&resp)
	}

	render.Status(r, resp.HTTPStatusCode)

	render.JSON(w, r, resp)
}
