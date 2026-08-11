// Package userapp maintains the app layer api for the user domain.
package userapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/api"
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

// create adds a new user to the system.
func (a *app) create(ctx context.Context, r *http.Request) (any, error) {
	var req NewUserRequest
	if err := decode(r, &req); err != nil {
		return nil, err
	}

	nu, err := toBusNewUser(req)
	if err != nil {
		return nil, err
	}

	usr, err := a.userBus.Create(ctx, nu)
	if err != nil {
		return nil, busError("create", err)
	}

	return api.NewStatus(http.StatusCreated, fromBusUserResponse(usr)), nil
}

// update modifies an existing user.
func (a *app) update(ctx context.Context, r *http.Request) (any, error) {
	usr, err := a.user(ctx, r)
	if err != nil {
		return nil, err
	}

	var req UpdateUserRequest
	if err := decode(r, &req); err != nil {
		return nil, err
	}

	uu, err := toBusUpdateUser(req)
	if err != nil {
		return nil, err
	}

	updUsr, err := a.userBus.Update(ctx, usr, uu)
	if err != nil {
		return nil, busError("update", err)
	}

	return fromBusUserResponse(updUsr), nil
}

// delete removes an existing user from the system.
func (a *app) delete(ctx context.Context, r *http.Request) (any, error) {
	usr, err := a.user(ctx, r)
	if err != nil {
		return nil, err
	}

	if err := a.userBus.Delete(ctx, usr); err != nil {
		return nil, busError("delete", err)
	}

	return api.NewStatus(http.StatusNoContent, nil), nil
}

// queryAll returns every user in the system.
func (a *app) queryAll(ctx context.Context, r *http.Request) (any, error) {
	usrs, err := a.userBus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(http.StatusInternalServerError, "queryall: %s", err)
	}

	return fromBusUsersResponse(usrs), nil
}

// queryByID returns the user identified by the specified id.
func (a *app) queryByID(ctx context.Context, r *http.Request) (any, error) {
	usr, err := a.user(ctx, r)
	if err != nil {
		return nil, err
	}

	return fromBusUserResponse(usr), nil
}

// user loads the user identified by the user_id path value.
func (a *app) user(ctx context.Context, r *http.Request) (userbus.User, error) {
	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil {
		return userbus.User{}, errs.Newf(http.StatusBadRequest, "invalid user id: %s", r.PathValue("user_id"))
	}

	usr, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		return userbus.User{}, busError("querybyid", err)
	}

	return usr, nil
}

// decode reads the JSON document in the request body into v.
func decode(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errs.Newf(http.StatusBadRequest, "invalid json: %s", err)
	}

	return nil
}

// busError maps a business layer error onto the matching HTTP status.
func busError(op string, err error) error {
	switch {
	case errors.Is(err, userbus.ErrNotFound):
		return errs.Newf(http.StatusNotFound, "%s: %s", op, err)

	case errors.Is(err, userbus.ErrUniqueEmail):
		return errs.Newf(http.StatusConflict, "%s: %s", op, err)

	default:
		return errs.Newf(http.StatusInternalServerError, "%s: %s", op, err)
	}
}
