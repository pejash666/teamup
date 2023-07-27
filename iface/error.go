package iface

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

func NewBackEndError(err error) *BackEndError {
	if err == nil {
		return &BackEndError{
			ErrNo:   0,
			ErrTips: "success",
		}
	}

}
