package login

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/golang-jwt/jwt/v4"
	"io/ioutil"
	"net/http"
	"teamup/model"
	"teamup/util"
	"time"
)

// Code2Session 通过前端获取的Code执行静默登录操作
func Code2Session(c *model.TeamUpContext, jsCode string) (*model.Code2Session, error) {
	if jsCode == "" {
		util.Logger.Println("jsCode is empty")
		return nil, errors.New("invalid jsCode")
	}

	// 获取access_token

	code2SessionUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?access_token=%s&appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.AccessToken, c.AppInfo.AppID, c.AppInfo.Secret, jsCode)
	resp, err := http.Get(code2SessionUrl)
	if err != nil {
		util.Logger.Printf("http.Get failed. err:%v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.Logger.Printf("ioutil.ReadAll failed, err:%v", err)
		return nil, err
	}
	c2s := &model.Code2Session{}
	err = sonic.Unmarshal(body, c2s)
	if err != nil {
		util.Logger.Printf("sonic.Unmarshal failed, err:%v", err)
		return nil, err
	}
	return c2s, nil
}

// CreateToken 对于登录态失效的用户，需要重新获取Token并下发给前端
func CreateToken(openID, sessionKey string) (string, error) {
	// 创建密钥
	secret := []byte("teamup's secret")
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
		util.Logger.Printf("token.SignedString failed, err:%v", err)
		return "", err
	}
	util.Logger.Printf("CreateToken success, res:%v", signedString)
	return signedString, nil
}
