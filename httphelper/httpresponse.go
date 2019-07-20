package httphelper

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

//组装http响应
func NewResponse(code int, message string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(message))),
	}
}
