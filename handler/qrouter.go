package handler

import (
	"fmt"
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
	context.Request.Host = u.Host
	context.Request.URL.Path = u.Path
	context.Request.URL.Scheme = "http"
	context.Request.RequestURI = ""
	context.Request.Header.Add("X-Forwarded-For", context.ClientIP)
	//referer := context.Request.Header.Get("Referer")
	//if referer != "" {
	//	uu, err := url.Parse(referer)
	//	if err == nil {
	//		ref := uu.Scheme + "://" + u.Host + uu.Path + "?" + uu.RawQuery
	//		//ref := strings.Replace(uu.RequestURI(), uu.Host, u.Host, 1)
	//
	//		context.Request.Header.Set("Referer", ref)
	//	}
	//}
	//proxy := func(_ *http.Request) (*url.URL, error) {
	//	return url.Parse("http://localhost:30006")
	//}
	//
	//client := &http.Client{
	//	Transport: &http.Transport{
	//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//		Proxy:           proxy,
	//	},
	//}
	//fmt.Println(context.Request.Host + context.Request.URL.RequestURI())
	resp, err := http.DefaultClient.Do(context.Request)
	if err != nil {
		fmt.Println("Handle", err)
		panic(err)
	}
	context.Response = resp
}
