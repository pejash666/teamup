package util

import (
	"fmt"
	"github.com/bytedance/sonic"
	"io/ioutil"
	"net/http"
	"teamup/model"
	"time"
)

// GetAccessToken 获取
func GetAccessToken(appID, appSecret string) (string, error) {
	// token有效期2小时，先从缓存取
	key := fmt.Sprintf("teamup_wechat_token")
	token, err := RedisGet(key)
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}
	// 缓存过期，重新获取
	getAccessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?appid=%s&secret=%s&js_code=%s&grant_type=client_credential",
		appID, appSecret)
	resp, err := http.Get(getAccessTokenUrl)
	if err != nil {
		Logger.Printf("http.Get failed. err:%v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger.Printf("ioutil.ReadAll failed, err:%v", err)
		return "", err
	}
	Logger.Printf("access_token resp:%v", string(body))

	at := &model.AccessTokenResp{}
	err = sonic.Unmarshal(body, at)
	if err != nil {
		Logger.Printf("sonic.Unmarshal failed, err:%v", err)
		return "", err
	}

	// 根据返回的过期时间写入缓存
	_ = RedisSet(key, at.AccessToken, time.Second*time.Duration(at.ExpiresIn))
	return at.AccessToken, nil
}
