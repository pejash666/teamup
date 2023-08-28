package info

import (
	"fmt"
	"github.com/bytedance/sonic"
	"io"
	"net/http"
	"strings"
	"teamup/model"
	"teamup/util"
)

func GetUserPhoneNumber(c *model.TeamUpContext, codeBody *model.GeneralCodeBody) (*model.PhoneInfoResp, error) {
	util.Logger.Printf("[GetUserPhoneNumber] starts, code:%v", codeBody.Code)

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", c.AccessToken)

	res, err := sonic.MarshalString(codeBody)
	if err != nil {
		util.Logger.Printf("[GetUserPhoneNumber] sonic marshall failed, err:%v", err)
		return nil, err
	}
	resp, err := http.Post(url, "application/json", strings.NewReader(res))
	if err != nil {
		util.Logger.Printf("[GetUserPhoneNumber] post failed, err:%v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.Logger.Printf("[GetUserPhoneNumber] read resp.Body failed, err:v", err)
		return nil, err
	}
	pI := &model.PhoneInfoResp{}
	err = sonic.Unmarshal(body, pI)
	if err != nil {
		util.Logger.Printf("[GetUserPhoneNumber] sonic unmarshall body failed, err:%v", err)
		return nil, err
	}
	util.Logger.Printf("[GetUserPhoneNumber] success, res:%+v", pI)
	return pI, nil
}
