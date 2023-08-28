package login

import (
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
