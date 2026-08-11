// Package db holds the database migrations for the project and provides them
// as an embedded filesystem so they ship inside the binary.
package db

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var migrations embed.FS

// Migrations returns the filesystem holding the goose migration files.
func Migrations() fs.FS {
	return migrations
}
