package handler

import (
	"github.com/qianyaozu/qgate/router"
	"net"
	"net/http"
	"strings"
)

type QRouter struct {
	//Handle(req *http.Request) (*http.Request, *http.Response)
}

func (qrouter *QRouter) Handle(r *http.Request) (*http.Request, *http.Response) {
	//hostL=r.Host
	//path := r.URL.Path
	//rawquery := r.URL.RawQuery
	//路由映射
	//u := "http://192.168.2.162:22019" + r.URL.Path
	var err error
	u, _, err := router.Conf.GetHttpLocation(ClientIP(r), r.Host, r.URL.Path)
	//fmt.Println(r.Host, r.URL.Path, u.Host, u.Path)
	r.URL.Host = u.Host
	r.URL.Path = u.Path
	r.URL.Scheme = "http"
	r.RequestURI = ""
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	return r, resp
}

//获取客户端IP
