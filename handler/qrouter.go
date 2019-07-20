package handler

import (
	"net/http"
)

type QRouter struct {
}

//路由转发
func (qrouter *QRouter) Handle(context *QContext) {
	//路由映射
	u, _, err := context.Server.GetHttpLocation(context.Request)
	if err != nil {
		panic(err)
	}
	context.Request.URL.Host = u.Host
	context.Request.URL.Path = u.Path
	context.Request.URL.Scheme = "http"
	context.Request.RequestURI = ""
	context.Request.Header.Add("X-Forwarded-For", context.ClientIP)
	client := &http.Client{}
	resp, err := client.Do(context.Request)
	if err != nil {
		panic(err)
	}
	context.Response = resp
}
