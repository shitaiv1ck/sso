package domain

import "github.com/shitaiv1ck/sso/internal/core/validation"

type App struct {
	ID   int
	Name string
}

func NewUnnamedApp(id int) App {
	return App{
		ID: id,
	}
}

func (a *App) Validate() error {
	return validation.ValidateID(a.ID)
}
