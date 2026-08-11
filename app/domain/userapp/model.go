package userapp

import (
	"time"

	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
)

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
