package login

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"io"
	"net/http"
	"teamup/model"
	"teamup/util"
)

// Code2Session 通过前端获取的Code执行静默登录操作
func Code2Session(c *model.TeamUpContext, jsCode string) (*model.Code2Session, error) {
	if jsCode == "" {
		util.Logger.Println("jsCode is empty")
		return nil, errors.New("invalid jsCode")
	}

	code2SessionUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?access_token=%s&appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.AccessToken, c.AppInfo.AppID, c.AppInfo.Secret, jsCode)
	resp, err := http.Get(code2SessionUrl)
	if err != nil {
		util.Logger.Printf("http.Get failed. err:%v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

// CheckDBSessionKeyExpired 用数据库的sessionKey + rawData 和 微信服务器做校验
func CheckDBSessionKeyExpired(c *model.TeamUpContext, sessionKey, rawData string) (bool, error) {
	if sessionKey == "" || rawData == "" {
		util.Logger.Println("session key,rawData or signature is empty")
		return true, errors.New("invalid params")
	}
	// 计算一个签名，用来和微信服务器校准数据库的sessionkey是否正确
	newHash := sha1.New()
	newHash.Write([]byte(rawData + sessionKey))
	newHashStr := hex.EncodeToString(newHash.Sum([]byte("")))
	checkSessionUrl := fmt.Sprintf("https://api.weixin.qq.com/wxa/checksession?access_token=%s&open_id=%s&signature=%s&sig_method=%s", c.AccessToken, c.BasicUser.OpenID, newHashStr, "hmac_sha256")
	resp, err := http.Get(checkSessionUrl)
	if err != nil {
		util.Logger.Printf("http.Get failed, err:%v", err)
		return true, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.Logger.Printf("io.ReadAll failed, err:%v", err)
		return true, err
	}
	base := &model.WechatBase{}
	err = sonic.Unmarshal(body, base)
	if err != nil {
		util.Logger.Printf("sonic.Unmarshal failed, err:%v", err)
		return true, err
	}
	util.Logger.Printf("base model:%v", base)
	return base.ErrCode != 0, nil
}

// CheckFrontEndSignature 服务端校准前端传来签名是否合法
func CheckFrontEndSignature(c *model.TeamUpContext, signature, sessionKey, rawData string) (bool, error) {
	newHash := sha1.New()
	newHash.Write([]byte(rawData + sessionKey))
	newHashStr := hex.EncodeToString(newHash.Sum([]byte("")))
	return newHashStr == signature, nil
}

func GetEncryptedData(c *model.TeamUpContext, sessionKey, encryptedData, iv string) ([]byte, error) {
	// 要解密的数据
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		util.Logger.Printf("[GetEncryptedData] DecodeString failed, err:%v", err)
		return nil, err
	}
	// aes的Key是session_key
	aesKey, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		util.Logger.Printf("[GetEncryptedData] DecodeString failed, err:%v", err)
		return nil, err
	}
	// 初始向量
	ivAfter, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		util.Logger.Printf("[GetEncryptedData] DecodeString failed, err:%v", err)
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		util.Logger.Printf("[GetEncryptedData] aes.NewCipher failed, err:%v", err)
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, ivAfter)
	res := make([]byte, len(data))
	blockMode.CryptBlocks(res, data)
	res, err = pkcs7UnPadding(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("aes data invalid")
	}
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}
