package iface

const (
	InvalidRequest = 1001

	InternalError = 1002
	WechatError   = 1003

	MysqlError = 1004

	ParamsError = 1005
)

type BackEndError struct {
	ErrNo   int32  `json:"err_no"`
	ErrTips string `json:"err_tips"`
}

func (e *BackEndError) Error() string {
	return e.ErrTips
}

func (e *BackEndError) ErrNumber() int32 {
	return e.ErrNo
}

func (e *BackEndError) WithErrNumber(errNo int32) *BackEndError {
	e.ErrNo = errNo
	return e
}

func (e *BackEndError) WithErrTips(tips string) *BackEndError {
	e.ErrTips = tips
	return e
}

func NewBackEndError(errCode int32, tips string) *BackEndError {
	return &BackEndError{
		ErrNo:   errCode,
		ErrTips: tips,
	}
}
