package util

import "encoding/json"

// 错误
type Err struct {
	Msg string
}

func (e *Err) Error() string {
	err, _ := json.Marshal(e)
	return string(err)
}

func NewError(msg string) *Err {
	return &Err{
		Msg: msg,
	}
}
