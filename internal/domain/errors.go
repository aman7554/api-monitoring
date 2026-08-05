package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden: insufficient permissions")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrOrgSlugExists     = errors.New("organization slug already exists")
	ErrProjectSlugExists = errors.New("project slug already exists")
	ErrInvalidInput      = errors.New("invalid input data")
)
