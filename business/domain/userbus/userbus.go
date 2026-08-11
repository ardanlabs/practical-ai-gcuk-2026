// Package userbus provides business access to the user domain.
package userbus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound    = errors.New("user not found")
	ErrUniqueEmail = errors.New("email is not unique")
)

// Storer defines the behavior this package needs from its storage layer.
type Storer interface {
	Create(ctx context.Context, usr User) (User, error)
	Update(ctx context.Context, usr User) (User, error)
	Delete(ctx context.Context, usr User) error
	QueryAll(ctx context.Context) ([]User, error)
	QueryByID(ctx context.Context, userID int64) (User, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log    *slog.Logger
	storer Storer
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *slog.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// Create adds a new user to the system.
func (b *Business) Create(ctx context.Context, nu NewUser) (User, error) {
	usr := User{
		Name:    nu.Name,
		Email:   nu.Email,
		Enabled: nu.Enabled,
	}

	usr, err := b.storer.Create(ctx, usr)
	if err != nil {
		return User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

// Update modifies information about a user.
func (b *Business) Update(ctx context.Context, usr User, uu UpdateUser) (User, error) {
	if uu.Name != nil {
		usr.Name = *uu.Name
	}

	if uu.Email != nil {
		usr.Email = *uu.Email
	}

	if uu.Enabled != nil {
		usr.Enabled = *uu.Enabled
	}

	usr, err := b.storer.Update(ctx, usr)
	if err != nil {
		return User{}, fmt.Errorf("update: %w", err)
	}

	return usr, nil
}

// Delete removes the specified user from the system.
func (b *Business) Delete(ctx context.Context, usr User) error {
	if err := b.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// QueryAll retrieves every user in the system.
func (b *Business) QueryAll(ctx context.Context) ([]User, error) {
	usrs, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return usrs, nil
}

// QueryByID finds the user by the specified user id.
func (b *Business) QueryByID(ctx context.Context, userID int64) (User, error) {
	usr, err := b.storer.QueryByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("querybyid: userID[%d]: %w", userID, err)
	}

	return usr, nil
}
