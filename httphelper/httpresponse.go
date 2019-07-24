package httphelper

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

//组装http响应
func NewResponse(statusCode int, message string) *http.Response {
	return &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(message))),
	}
}
