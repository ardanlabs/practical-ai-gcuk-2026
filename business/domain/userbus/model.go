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
