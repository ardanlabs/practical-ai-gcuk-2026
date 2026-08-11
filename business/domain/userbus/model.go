package userbus

import (
	"time"

	"github.com/ardanlabs/practical-ai-gcuk-2026/business/types/mail"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/types/name"
)

// User represents information about an individual user.
type User struct {
	ID          int64
	Name        name.Name
	Email       mail.Address
	Enabled     bool
	DateCreated time.Time
	DateUpdated time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name    name.Name
	Email   mail.Address
	Enabled bool
}

// UpdateUser contains information needed to update a user. A nil field
// means the existing value is left untouched.
type UpdateUser struct {
	Name    *name.Name
	Email   *mail.Address
	Enabled *bool
}
