package authredis

import (
	"context"
	"fmt"

	"github.com/shitaiv1ck/sso/internal/core/domain"
	"github.com/shitaiv1ck/sso/internal/core/repository/redis"
)

type AuthRedis struct {
	store redis.Redis
}

func NewAuthRedis(store redis.Redis) *AuthRedis {
	return &AuthRedis{
		store: store,
	}
}

func (r *AuthRedis) RevokeJWT(ctx context.Context, jwt domain.Token) error {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	key := fmt.Sprintf("blacklist:%s", jwt.JTI)

	restult := r.store.Set(ctx, key, "revoked", jwt.TTL)
	if err := restult.Err(); err != nil {
		return err
	}

	return nil
}
