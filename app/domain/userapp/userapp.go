// Package userapp maintains the app layer api for the user domain.
package userapp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/errs"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
)

type app struct {
	userBus *userbus.Business
}

func newApp(userBus *userbus.Business) *app {
	return &app{
		userBus: userBus,
	}
}

// queryAll returns every user in the system.
func (a *app) queryAll(ctx context.Context, r *http.Request) (any, error) {
	usrs, err := a.userBus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(http.StatusInternalServerError, "queryall: %s", err)
	}

	return fromBusUsersResponse(usrs), nil
}
