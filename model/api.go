package model

type APIOption struct {
	NeedLoginStatus    bool `json:"need_login_status"`    // 是否需要带着登录态
	NeedAdminClearance bool `json:"need_admin_clearance"` // 是否需要管理员权限
	HackLogic          bool `json:"hack_logic"`           // hack逻辑，不检查wechat_token
}

type LogInfo struct {
	FuncName string `json:"func_name"`
	Location string `json:"location"`
}

type GeneralCodeBody struct {
	Code string `json:"code"`
}
