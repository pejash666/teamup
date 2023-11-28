package model

type WechatBase struct {
	ErrCode int32  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type Code2Session struct {
	*WechatBase
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
}

type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type PhoneInfoResp struct {
	*WechatBase
	PhoneInfo *PhoneInfo `json:"phone_info"`
}

type PhoneInfo struct {
	PhoneNumber     string     `json:"phoneNumber"`
	PurePhoneNumber string     `json:"purePhoneNumber"`
	CountryCode     string     `json:"countryCode"`
	WaterMark       *Watermark `json:"watermark"`
}

type Watermark struct {
	TimeStamp int64  `json:"timeStamp"`
	AppID     string `json:"appid"`
}

type WechatUserInfo struct {
	NickName  string `json:"nick_name"`
	AvatarUrl string `json:"avatar_url"`
}
