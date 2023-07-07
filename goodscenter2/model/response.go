package model

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func SuccessResponse(data any) *Response {
	return &Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}
