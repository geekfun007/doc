package handler

import "byte.dance/pkg/errno"

type Response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func resp(e errno.Errno, data interface{}) *Response {
	return &Response{
		Code:    e.Code,
		Message: e.Message,
		Data:    data,
	}
}
