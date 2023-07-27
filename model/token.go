package model

import "github.com/golang-jwt/jwt/v4"

type TokenClaims struct {
	*jwt.RegisteredClaims

	OpenID     string `json:"open_id"`
	SessionKey string `json:"session_key"`
}
