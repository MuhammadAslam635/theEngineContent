package helpers

import (
	"errors"
	"time"

	"backend-go/internal/dto"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a new JWT token for a user
func GenerateToken(secret string, userID int32, email, username, utype string, duration time.Duration) (string, error) {
	claims := dto.TokenClaims{
		UserID:   userID,
		Email:    email,
		Username: username,
		UType:    utype,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// DecodeToken validates and parses a JWT token
func DecodeToken(tokenString string, secret string) (*dto.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*dto.TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
