package dto

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	UserID   int32  `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	UType    string `json:"utype"`
	jwt.RegisteredClaims
}
