package acckafka

type UserUpdatedDTO struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}
