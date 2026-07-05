package authredis

import (
	"context"
	"fmt"

	"github.com/shitaiv1ck/sso/internal/core/domain"
	"github.com/shitaiv1ck/sso/internal/core/repository/redis"
)

type AuthRepository struct {
	store redis.Redis
}

func NewAuthRep(store redis.Redis) *AuthRepository {
	return &AuthRepository{
		store: store,
	}
}

func (r *AuthRepository) RevokeJWT(ctx context.Context, jwt domain.Token) error {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	key := fmt.Sprintf("blacklist:%s", jwt.JTI)

	restult := r.store.Set(ctx, key, "revoked", jwt.TTL)
	if err := restult.Err(); err != nil {
		return err
	}

	return nil
}
