package main

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

//go:embed migrations/*.sql
var migrationsFS embed.FS

// GetDistFS returns the embedded dist filesystem, stripped of the dist prefix
func GetDistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

// GetMigrationsFS returns the embedded migrations filesystem, stripped of the migrations prefix
func GetMigrationsFS() (fs.FS, error) {
	return fs.Sub(migrationsFS, "migrations")
}
