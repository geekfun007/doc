package errno

import "fmt"

type Errno struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (e Errno) Error() string {
	return fmt.Sprintf("errno %d: %s", e.Code, e.Message)
}

func New(code int32, msg string) Errno {
	return Errno{Code: code, Message: msg}
}

var (
	Success        = New(0, "success")
	ParamErr       = New(10001, "invalid parameter")
	ServiceErr     = New(10002, "service internal error")
	AuthErr        = New(10003, "authentication failed")
	UserNotFound   = New(20001, "user not found")
	UserExists     = New(20002, "user already exists")
	ArticleNotFound = New(30001, "article not found")
)
