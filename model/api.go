package model

type APIOption struct {
	RequireMerchantUser bool `json:"require_merchant_user"` // 是否校验商家
	NeedLoginStatus     bool `json:"need_login_status"`     // 是否需要带着登录态
}

type LogInfo struct {
	FuncName string `json:"func_name"`
	Location string `json:"location"`
}

type GeneralCodeBody struct {
	Code string `json:"code"`
}
