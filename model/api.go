package model

type APIOption struct {
	RequireMerchantUser bool `json:"require_merchant_user"` // 是否校验商家
}

type LogInfo struct {
	FuncName string `json:"func_name"`
	Location string `json:"location"`
}

type UserLoginBody struct {
	Code string `json:"code"`
}
