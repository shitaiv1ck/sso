package authsrvc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
)

type AuthService struct {
	rep AuthRepository
}

type AuthRepository interface {
	SaveUser(ctx context.Context, user domain.User) (int, error)
	FindUser(ctx context.Context, user domain.User) (domain.User, error)
	FindApp(ctx context.Context, app domain.App) (domain.App, error)
}

func NewAuthService(rep AuthRepository) *AuthService {
	return &AuthService{
		rep: rep,
	}
}

func (s *AuthService) Register(ctx context.Context, user domain.User) (int, error) {
	if err := user.Validate(); err != nil {
		return -1, fmt.Errorf("failed to validate user: %w", err)
	}

	if err := user.HashingPassword(); err != nil {
		return -1, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := s.rep.SaveUser(ctx, user)
	if err != nil {
		return -1, fmt.Errorf("failed to save user: %w", err)
	}

	return userID, nil
}

func (s *AuthService) Login(ctx context.Context, user domain.User, app domain.App) (string, error) {
	if err := user.Validate(); err != nil {
		return "", fmt.Errorf("failed to validate user: %w", err)
	}

	if err := app.Validate(); err != nil {
		return "", fmt.Errorf("failed to validate app: %w", err)
	}

	foundUser, err := s.rep.FindUser(ctx, user)
	if err != nil {
		return "", errs.ErrInvalidCredentials
	}

	if !foundUser.ComparePassword(user.Password) {
		return "", errs.ErrInvalidCredentials
	}

	foundApp, err := s.rep.FindApp(ctx, app)
	if err != nil {
		return "", errs.ErrInvalidCredentials
	}

	token, err := newToken(user, foundApp)
	if err != nil {
		return "", err
	}

	return token, nil
}

func newToken(user domain.User, app domain.App) (string, error) {
	duration, err := time.ParseDuration(os.Getenv("JWT_TTL"))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(duration).Unix(),
	})

	signingKey, err := getJWTKey(app.Name)
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getJWTKey(appName string) (string, error) {
	keyName := fmt.Sprintf("JWT_%s_KEY", strings.ToUpper(appName))

	jwtKey := os.Getenv(keyName)
	if jwtKey == "" {
		return "", errs.ErrKeyNotConfigured
	}

	return jwtKey, nil
}
