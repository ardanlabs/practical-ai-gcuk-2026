// Package userbus provides business access to the user domain.
package userbus

import (
	"context"
	"fmt"
	"log/slog"
)

// Storer defines the behavior this package needs from its storage layer.
type Storer interface {
	QueryAll(ctx context.Context) ([]User, error)
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

// QueryAll retrieves every user in the system.
func (b *Business) QueryAll(ctx context.Context) ([]User, error) {
	usrs, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return usrs, nil
}
