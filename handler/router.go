package handler

import (
	"fmt"
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
	fmt.Println(r.Host, r.URL.Path, u.Host, u.Path)
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
func ClientIP(r *http.Request) string {
	if clientIPs := r.Header["X-Forwarded-For"]; len(clientIPs) > 0 {
		clientIP := clientIPs[0]
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if len(clientIP) > 0 {
			return clientIP
		}

	}
	if clientIPs := r.Header["X-Real-Ip"]; len(clientIPs) > 0 {
		return clientIPs[0]
	}
	if clientIPs := r.Header["X-Appengine-Remote-Addr"]; len(clientIPs) > 0 {
		return clientIPs[0]
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
