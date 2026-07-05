package authkafka

type UserCreatedDTO struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}
