// Package checkapp maintains the app layer api for the check domain.
package checkapp

import (
	"context"
	"net/http"
)

type app struct{}

func newApp() *app {
	return &app{}
}

// liveness returns a 200 to signal the service is up and running.
func (a *app) liveness(ctx context.Context, r *http.Request) (any, error) {
	return status{Status: "OK"}, nil
}
