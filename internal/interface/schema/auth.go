package schema

import (
	"github.com/golang-jwt/jwt/v5"
)

// Authenticator が発行するJWTのclaims。Subject は Discord のユーザーID。
// UserID は JWT には含まれず、Backend のミドルウェアで設定する。
type JWTCustomClaims struct {
	UserID   string `json:"-"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}
