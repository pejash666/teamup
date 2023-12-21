package notice

import (
	"fmt"
	"github.com/bytedance/sonic"
	"io"
	"net/http"
	"strings"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type NoticeData map[string]struct {
	Value string `json:"value"`
}

// InformUser 在微信里通知用户进展
// templateID
func InformUser(c *model.TeamUpContext, templateID string, data map[string]string, page string, toUser string, lang string) error {
	Url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", c.AccessToken)
	templateData := ""
	if len(data) > 0 {
		nd := toNoticeData(data)
		str, err := sonic.MarshalString(nd)
		if err != nil {
			util.Logger.Printf("[InformUser] marshal data failed, err:%v", err)
			return iface.NewBackEndError(iface.InternalError, err.Error())
		}
		templateData = str
	}
	// 构建请求Body
	bodyM := make(map[string]string)
	bodyM["template_id"] = templateID
	bodyM["page"] = page
	bodyM["touser"] = toUser
	bodyM["data"] = templateData
	bodyM["miniprogram_state"] = "formal"
	bodyM["lang"] = c.Language

	reqBody, err := sonic.MarshalString(bodyM)
	if err != nil {
		util.Logger.Printf("[InformUser] marshall body failed, err:%v", err)
		return iface.NewBackEndError(iface.InternalError, err.Error())
	}

	// 1. 有用户加入了创建的活动，活动主会收到提示
	resp, err := http.Post(Url, "application/json", strings.NewReader(reqBody))
	if err != nil {
		util.Logger.Printf("[InformUser] query wechat's url failed, err:%v", err)
		return iface.NewBackEndError(iface.InternalError, err.Error())
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		util.Logger.Printf("[InformUser] read resp.Body failed, err:v", err)
		return iface.NewBackEndError(iface.InternalError, err.Error())
	}
	baseResp := &model.WechatBase{}
	err = sonic.Unmarshal(respBody, baseResp)
	if err != nil {
		util.Logger.Printf("[InformUser] unmarshal failed, err:%v", err)
		return iface.NewBackEndError(iface.InternalError, err.Error())
	}
	if baseResp.ErrCode != 0 {
		util.Logger.Printf("[InformUser] errCode is not zero, code:%v", baseResp.ErrCode)
		return iface.NewBackEndError(iface.WechatError, baseResp.ErrMsg)
	}
	return nil
}

func toNoticeData(data map[string]string) NoticeData {
	res := make(NoticeData)
	for k, v := range data {
		res[k] = struct {
			Value string `json:"value"`
		}{v}
	}
	return res
}
