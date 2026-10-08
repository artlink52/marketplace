package lib

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID string
	Role   string
}

func NewToken(c Claims, secret string, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = c.UserID
	claims["role"] = c.Role
	claims["exp"] = time.Now().Add(ttl).Unix()

	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString, secret string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, ErrInvalidToken
	}

	userID, ok := mapClaims["uid"].(string)
	if !ok {
		return Claims{}, ErrInvalidToken
	}

	role, ok := mapClaims["role"].(string)
	if !ok {
		return Claims{}, ErrInvalidToken
	}

	return Claims{UserID: userID, Role: role}, nil
}
