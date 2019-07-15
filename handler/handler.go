package handler

import "net/http"

type ReqHandler interface {
	Handle(req *http.Request) (*http.Request, *http.Response)
}

type RespHandler interface {
	Handle(req *http.Request, resp *http.Response) *http.Response
}
