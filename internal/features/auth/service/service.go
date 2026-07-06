package authsrvc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
	"github.com/shitaiv1ck/sso/internal/core/validation"
)

type AuthService struct {
	config Config

	pg AuthPostgres
	tx postgres.TransactionManager

	redis AuthRedis

	kafka AuthKafka
}

type AuthPostgres interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	FindUser(ctx context.Context, user domain.User) (domain.User, error)
	FindApp(ctx context.Context, app domain.App) (domain.App, error)

	SaveSession(ctx context.Context, session domain.Session) (string, error)
	SaveTXSession(ctx context.Context, tx postgres.SQLExecuter, session domain.Session) (string, error)
	DeleteSession(ctx context.Context, refreshToken string) error
	DeleteTXSession(ctx context.Context, tx postgres.SQLExecuter, refreshToken string) (domain.Session, domain.User, error)
}

type AuthRedis interface {
	RevokeJWT(ctx context.Context, jwt domain.Token) error
}

type AuthKafka interface {
	EventUserRegistered(ctx context.Context, user domain.User) error
}

func NewAuthService(
	config Config,
	pg AuthPostgres,
	tx postgres.TransactionManager,
	redis AuthRedis,
	kafka AuthKafka,
) *AuthService {
	return &AuthService{
		config: config,
		pg:     pg,
		tx:     tx,
		redis:  redis,
		kafka:  kafka,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (int, error) {
	user := domain.NewUnknownUser(email, password)
	if err := user.Validate(); err != nil {
		return -1, fmt.Errorf("failed to validate user: %w", err)
	}

	if err := user.HashingPassword(); err != nil {
		return -1, fmt.Errorf("failed to hash password: %w", err)
	}

	savedUser, err := s.pg.SaveUser(ctx, user)
	if err != nil {
		return -1, fmt.Errorf("failed to save user: %w", err)
	}

	s.kafka.EventUserRegistered(ctx, savedUser)

	return savedUser.ID, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string, appID int) (string, string, error) {
	user := domain.NewUnknownUser(email, password)
	if err := user.Validate(); err != nil {
		return "", "", fmt.Errorf("failed to validate user: %w", err)
	}

	app := domain.NewUnnamedApp(appID)
	if err := app.Validate(); err != nil {
		return "", "", fmt.Errorf("failed to validate app: %w", err)
	}

	foundUser, err := s.pg.FindUser(ctx, user)
	if err != nil {
		return "", "", errs.ErrInvalidCredentials
	}

	if !foundUser.ComparePassword(user.Password) {
		return "", "", errs.ErrInvalidCredentials
	}

	foundApp, err := s.pg.FindApp(ctx, app)
	if err != nil {
		return "", "", errs.ErrInvalidCredentials
	}

	accessToken, err := s.newAccessToken(foundUser.ID, foundApp.Name)
	if err != nil {
		return "", "", err
	}

	token, err := generateToken()
	if err != nil {
		return "", "", err
	}

	session, err := domain.NewSession(token, foundUser.ID, s.config.SessionTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	refreshToken, err := s.pg.SaveSession(ctx, session)
	if err != nil {
		return "", "", fmt.Errorf("failed to save session: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, appID int) (string, string, error) {
	if err := validation.ValidateRefreshToken(refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if err := validation.ValidateID(appID); err != nil {
		return "", "", fmt.Errorf("failed to validate app ID: %w", err)
	}

	app := domain.NewUnnamedApp(appID)
	foundApp, err := s.pg.FindApp(ctx, app)
	if err != nil {
		return "", "", fmt.Errorf("failed to find app: %w", err)
	}

	tx, err := s.tx.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	deletedSession, user, err := s.pg.DeleteTXSession(ctx, tx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to delete session: %w", errs.ErrRefSession)
	}

	if deletedSession.IsExpired() {
		return "", "", fmt.Errorf("refresh token is expired: %w", errs.ErrRefSession)
	}

	token, err := generateToken()
	if err != nil {
		return "", "", err
	}

	session, err := domain.NewSession(token, user.ID, s.config.SessionTTL)
	if err != nil {
		return "", "", errs.ErrRefSession
	}

	refreshToken, err = s.pg.SaveTXSession(ctx, tx, session)
	if err != nil {
		return "", "", fmt.Errorf("failed to save session: %w", errs.ErrRefSession)
	}

	accessToken, err := s.newAccessToken(user.ID, foundApp.Name)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) newAccessToken(userID int, appName string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JwtTTL)),
		ID:        uuid.NewString(),
	})

	signingKey, err := getJWTKey(appName)
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

var keyCache = sync.Map{}

func getJWTKey(appName string) (string, error) {
	keyName := fmt.Sprintf("JWT_%s_KEY", strings.ToUpper(appName))

	var jwtKey string
	if key, ok := keyCache.Load(appName); ok {
		jwtKey = key.(string)
	} else {
		jwtKey = os.Getenv(keyName)
		if jwtKey == "" {
			return "", errs.ErrKeyNotConfigured
		}

		keyCache.Store(appName, jwtKey)
	}

	return jwtKey, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, accessToken string) error {
	if err := validation.ValidateRefreshToken(refreshToken); err != nil {
		return fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if err := s.pg.DeleteSession(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	token, err := parseToken(accessToken)
	if err == nil {
		s.redis.RevokeJWT(ctx, token)
	}

	return nil
}

func parseToken(accessToken string) (domain.Token, error) {
	parsedToken, _, err := jwt.NewParser().ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return domain.Token{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return domain.Token{}, fmt.Errorf("failed to get claims: %w", errs.ErrInvalidJWT)
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return domain.Token{}, fmt.Errorf("failed to get 'jti': %w", errs.ErrInvalidJWT)
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return domain.Token{}, fmt.Errorf("failed to get 'exp': %w", errs.ErrInvalidJWT)
	}

	expAt := time.Unix(int64(exp), 0)

	if expAt.Before(time.Now()) {
		return domain.Token{}, fmt.Errorf("token expired: %w", errs.ErrInvalidJWT)
	}

	token := domain.NewToken(jti, time.Until(expAt))

	return token, nil
}
