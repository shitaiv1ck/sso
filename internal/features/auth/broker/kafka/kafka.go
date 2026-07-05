package authkafka

import (
	"context"

	"github.com/shitaiv1ck/sso/internal/core/broker/kafka"
	"github.com/shitaiv1ck/sso/internal/core/domain"
)

type AuthKafka struct {
	producer kafka.Producer
}

func NewAuthKafka(producer kafka.Producer) *AuthKafka {
	return &AuthKafka{
		producer: producer,
	}
}

func (k *AuthKafka) EventUserCreated(ctx context.Context, user domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, k.producer.GetTimeout())
	defer cancel()

	msg := UserCreatedDTO{
		UserID: user.ID,
		Email:  user.Email,
	}

	record, err := k.producer.NewRecord("user.created", msg)
	if err != nil {
		return err
	}

	result := k.producer.ProduceSync(ctx, record)
	if result.FirstErr() != nil {
		return result.FirstErr()
	}

	return nil
}
