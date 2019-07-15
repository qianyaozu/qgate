package server

import (
	"fmt"
	"github.com/qianyaozu/qgate/handler"
	"github.com/qianyaozu/qlog"
	"io"
	"net/http"
	"os"
	"time"
)

type QReverseServer struct {
	requestHandlers  []handler.ReqHandler
	responseHandlers []handler.RespHandler
}

type QRequestLog struct {
	StartTime    time.Time     //开始统计时间
	EndTime      time.Time     //结束统计时间
	Latency      time.Duration //请求耗时
	Host         string
	OriginPath   string //原始请求路径
	Path         string //实际转发路径
	RawQuery     string //查询条件
	Header       string //请求header
	RequestBody  string //请求body
	ResponseBody string //响应内容

	ClientIP   string //访问地址IP
	Method     string //请求方式
	StatusCode int    //请求结果编码
	Error      interface{}
}

func NewProxyServer() *QReverseServer {
	return &QReverseServer{
		requestHandlers:  make([]handler.ReqHandler, 0),
		responseHandlers: make([]handler.RespHandler, 0),
	}
}

func (server *QReverseServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var log = QRequestLog{
		StartTime:  time.Now(),
		Host:       r.Host,
		OriginPath: r.URL.Path,
		Path:       "",
		RawQuery:   r.URL.RawQuery,
		ClientIP:   handler.ClientIP(r),
		Method:     r.Method,
		StatusCode: 200,
		//Header:r.Header
	}
	defer func() {
		log.EndTime = time.Now()
		if err := recover(); err != nil {
			w.WriteHeader(500)
			log.Error = err
			log.StatusCode = 500
		}
		log.Latency = log.EndTime.Sub(log.StartTime)
		l := fmt.Sprintf("[QGate]  %v | %v | %v | %v | %v | %v | %v \n",
			log.StatusCode, log.Latency, log.ClientIP, log.Method, log.Host, log.OriginPath, log.Path)
		fmt.Fprintf(os.Stdout, l)
		qlog.Trace("access", l)
	}()
	//过滤请求
	r, resp := server.filterRequest(r)
	log.Path = r.URL.Path
	if resp == nil {
		//如果没有返回数据，则统一返回404
		w.WriteHeader(404)
		return
	}
	origBody := resp.Body
	//过滤返回数据
	resp = server.filterResponse(r, resp)
	defer origBody.Close()
	if origBody != resp.Body {
		resp.Header.Del("Content-Length")
	}

	server.copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
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
func (proxy *QReverseServer) filterRequest(oldreq *http.Request) (req *http.Request, resp *http.Response) {
	req = oldreq
	for _, h := range proxy.requestHandlers {
		req, resp = h.Handle(oldreq)
		if resp != nil {
			break
		}
	}
	return
}

//过滤response Handler
func (proxy *QReverseServer) filterResponse(oldreq *http.Request, oleresp *http.Response) (resp *http.Response) {
	resp = oleresp
	for _, h := range proxy.responseHandlers {
		resp = h.Handle(oldreq, oleresp)
	}
	return
}

//插入request处理中间件
func (proxy *QReverseServer) UseRequestHandler(handle handler.ReqHandler) {
	proxy.requestHandlers = append(proxy.requestHandlers, handle)

}

//插入response处理中间件
func (proxy *QReverseServer) UseResponseHandler(handle handler.RespHandler) {
	proxy.responseHandlers = append(proxy.responseHandlers, handle)
}
