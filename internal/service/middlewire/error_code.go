package middlewire

type ErrorCode int32

const (
	ParamErrorCode ErrorCode = 0
	SysErrorCode             = 1
)

var HttpErrorCodeDetail = map[ErrorCode]string{
	ParamErrorCode: "参数错误",
	SysErrorCode:   "系统错误",
}

func (e ErrorCode) String() string {
	msg, ok := HttpErrorCodeDetail[e]
	if ok {
		return msg
	}
	return "UNKNOW"
}
