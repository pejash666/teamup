package util

import (
	"github.com/golang-jwt/jwt/v4"
	"teamup/constant"
	"teamup/model"
	"time"
)

// CreateJWTToken 对于登录态失效的用户，需要重新获取Token并下发给前端
func CreateJWTToken(c *model.TeamUpContext, openID, sessionKey string) (string, error) {
	// 创建密钥
	secret := []byte(constant.JsonWebTokenSecret)
	// 创建Claims
	tokenClaims := model.TokenClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			Issuer:    "team_up_server",
			Subject:   "mini_program_token",
			Audience:  jwt.ClaimStrings{"mini_program_fe"},
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		OpenID:     openID,
		SessionKey: sessionKey,
	}
	// 创建Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	// 签名
	signedString, err := token.SignedString(secret)
	if err != nil {
		Logger.Printf("token.SignedString failed, err:%v", err)
		return "", err
	}
	Logger.Printf("CreateJWTToken success, res:%v", signedString)
	return signedString, nil
}

func ParseJWTToken(tokenStr string) (*model.TokenClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenStr, &model.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(constant.JsonWebTokenSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*model.TokenClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
