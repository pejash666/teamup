package model

type BackEndResp struct {
	ErrNo   int32       `json:"err_no"`
	ErrTips string      `json:"err_tips"`
	Data    interface{} `json:"data"`
}
