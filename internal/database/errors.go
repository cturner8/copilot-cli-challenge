package database

import (
	"errors"
)

var (
	// ErrNotFound indicates the requested entity was not found
	ErrNotFound = errors.New("entity not found")

	// ErrDuplicate indicates a unique constraint violation
	ErrDuplicate = errors.New("duplicate entity")

	// ErrForeignKey indicates a foreign key constraint violation
	ErrForeignKey = errors.New("foreign key constraint violation")
)
