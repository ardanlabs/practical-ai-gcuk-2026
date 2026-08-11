package userdb

import (
	"fmt"
	"time"

	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/types/mail"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/types/name"
)

type dbUser struct {
	ID          int64     `db:"user_id"`
	Name        string    `db:"name"`
	Email       string    `db:"email"`
	Enabled     bool      `db:"enabled"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toDBUser(bus userbus.User) dbUser {
	db := dbUser{
		ID:          bus.ID,
		Name:        bus.Name.String(),
		Email:       bus.Email.String(),
		Enabled:     bus.Enabled,
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}

	return db
}

func toBusUser(db dbUser) (userbus.User, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse name: %w", err)
	}

	addr, err := mail.Parse(db.Email)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse email: %w", err)
	}

	bus := userbus.User{
		ID:          db.ID,
		Name:        nme,
		Email:       addr,
		Enabled:     db.Enabled,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toBusUsers(dbs []dbUser) ([]userbus.User, error) {
	bus := make([]userbus.User, len(dbs))

	for i, db := range dbs {
		var err error
		if bus[i], err = toBusUser(db); err != nil {
			return nil, err
		}
	}

	return bus, nil
}
