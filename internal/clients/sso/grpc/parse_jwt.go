package grpc

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// ParseUserIDFromToken извлекает user_id из jwt токена
func ParseUserIDFromToken(tokenString string, secret string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["uid"].(float64); ok {
			return int64(uid), nil
		}
		return 0, errors.New("uid not found in token")
	}
	return 0, errors.New("invalid token claims")
}

// ParseEmailFromToken извлекает email из jwt токена
func ParseEmailFromToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, ok := claims["email"].(string); ok {
			return email, nil
		}
		return "", errors.New("email not found in token")
	}
	return "", errors.New("invalid token claims")
}
