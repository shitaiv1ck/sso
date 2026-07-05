package domain

import "time"

type Token struct {
	JTI string
	TTL time.Duration
}

func NewToken(jti string, ttl time.Duration) Token {
	return Token{
		JTI: jti,
		TTL: ttl,
	}
}
