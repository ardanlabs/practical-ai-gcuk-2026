// Package helloapp maintains the app layer api for the hello domain.
package helloapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/errs"
)

type app struct{}

func newApp() *app {
	return &app{}
}

// hello greets the user provided in the route.
func (a *app) hello(ctx context.Context, r *http.Request) (any, error) {
	user := r.PathValue("user")
	if user == "" {
		return nil, errs.Newf(http.StatusBadRequest, "user is required")
	}

	return greetingResponse{Message: fmt.Sprintf("Hello, %s", user)}, nil
}
