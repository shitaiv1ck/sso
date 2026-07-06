package domain

import (
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/validation"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Email        string
	Password     string
	PasswordHash string
}

func NewUnknownUser(email string, password string) User {
	return User{
		ID:       -1,
		Email:    email,
		Password: password,
	}
}

func (u *User) Validate() error {
	if err := validation.ValidateEmail(u.Email); err != nil {
		return err
	}

	if err := validation.ValidatePassword(u.Password); err != nil {
		return err
	}

	return nil
}

func (u *User) HashingPassword() error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(passwordHash)

	return nil
}

func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))

	return err == nil
}

func (u *User) ChangePassword(oldPassword string, newPassword string) error {
	if err := validation.ValidatePassword(oldPassword); err != nil {
		return err
	}

	if err := validation.ValidatePassword(newPassword); err != nil {
		return err
	}

	if oldPassword == newPassword {
		return errs.ErrInvalidArg
	}

	if !u.ComparePassword(oldPassword) {
		return errs.ErrInvalidArg
	}

	u.Password = newPassword
	if err := u.HashingPassword(); err != nil {
		return err
	}

	return nil
}

func (u *User) ChangeEmail(newEmail string) error {
	if err := validation.ValidateEmail(newEmail); err != nil {
		return errs.ErrInvalidArg
	}

	if u.Email == newEmail {
		return errs.ErrInvalidArg
	}

	u.Email = newEmail

	return nil
}
