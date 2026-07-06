package acckafka

import (
	"context"

	"github.com/shitaiv1ck/sso/internal/core/client/kafka"
	"github.com/shitaiv1ck/sso/internal/core/domain"
)

type AccountKafka struct {
	producer kafka.Producer
}

func NewAccountKafka(producer kafka.Producer) *AccountKafka {
	return &AccountKafka{
		producer: producer,
	}
}

func (k *AccountKafka) EventUserUpdated(ctx context.Context, user domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, k.producer.GetTimeout())
	defer cancel()

	msg := UserUpdatedDTO{
		UserID: user.ID,
		Email:  user.Email,
	}

	record, err := k.producer.NewRecord("user.updated", msg)
	if err != nil {
		return err
	}

	result := k.producer.ProduceSync(ctx, record)
	if result.FirstErr() != nil {
		return result.FirstErr()
	}

	return nil
}
