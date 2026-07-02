package validation

import (
	"fmt"
	"regexp"
	"strings"

	errs "github.com/shitaiv1ck/sso/internal/core/errors"
)

var emailRg = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email must be non empty: %w", errs.ErrInvalidArg)
	}

	if !emailRg.MatchString(email) {
		return fmt.Errorf("invalid email format: %w", errs.ErrInvalidArg)
	}

	return nil
}

var passwordRg = regexp.MustCompile(`^[a-zA-Z0-9+-_!&*$?/\><]+$`)

func ValidatePassword(password string) error {
	trimmedPassword := strings.TrimSpace(password)
	if trimmedPassword == "" {
		return fmt.Errorf("password must be non empty: %w", errs.ErrInvalidArg)
	}

	lenTrimmedPassword := len([]rune(trimmedPassword))
	if lenTrimmedPassword < 8 || lenTrimmedPassword > 100 {
		return fmt.Errorf("password length must be between 8 and 100: %w", errs.ErrInvalidArg)
	}

	if !passwordRg.MatchString(password) {
		return fmt.Errorf("invalid password format: %w", errs.ErrInvalidArg)
	}

	return nil
}

func ValidateID(id int) error {
	if id <= 0 {
		return fmt.Errorf("ID must be positive: %w", errs.ErrInvalidArg)
	}

	return nil
}
