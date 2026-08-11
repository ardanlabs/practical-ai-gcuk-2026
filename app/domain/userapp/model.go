package userapp

import (
	"time"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/errs"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/types/mail"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/types/name"
)

// NewUserRequest defines the data needed to add a new user, using primitive
// types only. When Enabled is not provided the user is enabled.
type NewUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Enabled *bool  `json:"enabled"`
}

// UpdateUserRequest defines the data needed to update a user. A nil field
// means the existing value is left untouched.
type UpdateUserRequest struct {
	Name    *string `json:"name"`
	Email   *string `json:"email"`
	Enabled *bool   `json:"enabled"`
}

// toBusNewUser converts the request into the business layer representation,
// parsing the strong types. All invalid fields are reported together.
func toBusNewUser(app NewUserRequest) (userbus.NewUser, error) {
	var fieldErrors errs.FieldErrors

	nme, err := name.Parse(app.Name)
	if err != nil {
		fieldErrors = append(fieldErrors, errs.FieldError{Field: "name", Err: err.Error()})
	}

	addr, err := mail.Parse(app.Email)
	if err != nil {
		fieldErrors = append(fieldErrors, errs.FieldError{Field: "email", Err: err.Error()})
	}

	if len(fieldErrors) > 0 {
		return userbus.NewUser{}, fieldErrors
	}

	enabled := true
	if app.Enabled != nil {
		enabled = *app.Enabled
	}

	bus := userbus.NewUser{
		Name:    nme,
		Email:   addr,
		Enabled: enabled,
	}

	return bus, nil
}

// toBusUpdateUser converts the request into the business layer representation,
// parsing the strong types. All invalid fields are reported together.
func toBusUpdateUser(app UpdateUserRequest) (userbus.UpdateUser, error) {
	var fieldErrors errs.FieldErrors

	var nme *name.Name
	if app.Name != nil {
		n, err := name.Parse(*app.Name)
		if err != nil {
			fieldErrors = append(fieldErrors, errs.FieldError{Field: "name", Err: err.Error()})
		} else {
			nme = &n
		}
	}

	var addr *mail.Address
	if app.Email != nil {
		a, err := mail.Parse(*app.Email)
		if err != nil {
			fieldErrors = append(fieldErrors, errs.FieldError{Field: "email", Err: err.Error()})
		} else {
			addr = &a
		}
	}

	if len(fieldErrors) > 0 {
		return userbus.UpdateUser{}, fieldErrors
	}

	bus := userbus.UpdateUser{
		Name:    nme,
		Email:   addr,
		Enabled: app.Enabled,
	}

	return bus, nil
}

// UserResponse represents information about an individual user at the API
// layer, using primitive types only.
type UserResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Enabled     bool   `json:"enabled"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

func fromBusUserResponse(bus userbus.User) UserResponse {
	return UserResponse{
		ID:          bus.ID,
		Name:        bus.Name.String(),
		Email:       bus.Email.String(),
		Enabled:     bus.Enabled,
		DateCreated: bus.DateCreated.Format(time.RFC3339),
		DateUpdated: bus.DateUpdated.Format(time.RFC3339),
	}
}

func fromBusUsersResponse(bus []userbus.User) []UserResponse {
	app := make([]UserResponse, len(bus))

	for i, b := range bus {
		app[i] = fromBusUserResponse(b)
	}

	return app
}
