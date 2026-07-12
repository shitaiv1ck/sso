package errs

import "errors"

var (
	ErrInvalidArg         = errors.New("invalid argument")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrKeyNotConfigured   = errors.New("key not configured")
	ErrNotFound           = errors.New("not found")
	ErrRefSession         = errors.New("invalid refresh session")
	ErrInvalidJWT         = errors.New("invalid JWT")
)
