package authkafka

type UserRegisteredDTO struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}
