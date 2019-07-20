package server

import (
	"fmt"
	"github.com/qianyaozu/qgate/handler"
	"github.com/qianyaozu/qgate/httphelper"
	"github.com/qianyaozu/qgate/router"
	"io"
	"net/http"
)

type QReverseServer struct {
	requestHandlers  []handler.QHandler
	responseHandlers []handler.QHandler
}

func NewProxyServer() *QReverseServer {
	return &QReverseServer{
		requestHandlers:  make([]handler.QHandler, 0),
		responseHandlers: make([]handler.QHandler, 0),
	}
}

func (server *QReverseServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//初始化请求信息
	context := handler.InitContext(r)
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(httphelper.INTERNAL_SERVER_ERROR)
			if context.Response == nil {
				context.Response = httphelper.NewResponse(httphelper.INTERNAL_SERVER_ERROR, "INTERNAL_SERVER_ERROR "+fmt.Sprint(err))
			} else {
				context.Response.StatusCode = httphelper.INTERNAL_SERVER_ERROR
			}
		}
		//处理日志信息
		handler.HandleLog(context)
	}()

	//根据配置信息获取路由信息
	sconfig, err := router.Conf.GetHttpServer(context.Request)
	if err != nil {
		//服务访问出错
		context.Response = httphelper.NewResponse(httphelper.BAD_GATEWAY, "BAD_GATEWAY "+err.Error())
		goto writeresponse
	}
	if sconfig.Return > 0 {
		//服务禁止访问
		context.Response = &http.Response{StatusCode: sconfig.Return}
		goto writeresponse
	}
	context.Server = sconfig

	//过滤请求
	server.filterRequest(context)
	if context.Response == nil {
		//如果没有返回数据，则统一返回404
		w.WriteHeader(httphelper.NOT_FOUND)
		return
	}
	//origBody := context.Request.Body
	//过滤返回数据
	server.filterResponse(context)
	//defer origBody.Close()
	//if origBody != context.Response.Body {
	//	context.Response.Header.Del("Content-Length")
	//}

writeresponse:
	server.copyHeaders(w.Header(), context.Response.Header)
	w.WriteHeader(context.Response.StatusCode)
	io.Copy(w, context.Response.Body)
	context.Response.Body.Close()
}

//复制response的Header
func (proxy *QReverseServer) copyHeaders(dst, src http.Header) {
	for k, _ := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

//过滤request Handler
func (proxy *QReverseServer) filterRequest(context *handler.QContext) {
	for _, h := range proxy.requestHandlers {
		h.Handle(context)
		if context.Response != nil {
			break
		}
	}
	return
}

//过滤response Handler
func (proxy *QReverseServer) filterResponse(context *handler.QContext) {
	for _, h := range proxy.responseHandlers {
		h.Handle(context)
	}
	return
}

//插入request处理中间件
func (proxy *QReverseServer) UseRequestHandler(handle handler.QHandler) {
	proxy.requestHandlers = append(proxy.requestHandlers, handle)

}

//插入response处理中间件
func (proxy *QReverseServer) UseResponseHandler(handle handler.QHandler) {
	proxy.responseHandlers = append(proxy.responseHandlers, handle)
}
