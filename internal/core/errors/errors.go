package errs

import "errors"

var (
	ErrInvalidArg         = errors.New("Invalid argument")
	ErrAlreadyExist       = errors.New("Already exist")
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrKeyNotConfigured   = errors.New("Key not configured")
	ErrNotFound           = errors.New("Not found")
)
