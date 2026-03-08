package jwtauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Config struct {
	Secret         string
	AccessTokenTTL time.Duration
}

type Provider struct {
	secret         []byte
	accessTokenTTL time.Duration
}

type claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func New(cfg Config) *Provider {
	return &Provider{
		secret:         []byte(cfg.Secret),
		accessTokenTTL: cfg.AccessTokenTTL,
	}
}

// IssueAccessToken signs a new JWT for the given userID.
func (p *Provider) IssueAccessToken(userID string) (string, error) {
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// ParseAccessToken validates the token and returns the userID embedded in it.
func (p *Provider) ParseAccessToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return p.secret, nil
	})
	if err != nil {
		return "", err
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	return c.UserID, nil
}
