package lib

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(userID, role, secret string, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = userID
	claims["role"] = role
	claims["exp"] = time.Now().Add(ttl).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
