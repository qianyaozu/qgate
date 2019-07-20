package handler

import (
	"fmt"
	"github.com/qianyaozu/qlog"
	"os"
	"time"
)

type QRequestLog struct {
	Latency      time.Duration //请求耗时
	Host         string
	OriginPath   string //原始请求路径
	Path         string //实际转发路径
	RawQuery     string //查询条件
	Header       string //请求header
	RequestBody  string //请求body
	ResponseBody string //响应内容

	ClientIP    string //访问地址IP
	Method      string //请求方式
	StatusCode  int    //请求结果编码
	ContentType string
	Error       interface{}
}

func HandleLog(context *QContext) {
	defer func() {

	}()
	var log = QRequestLog{

		Host:       context.Request.Host,
		OriginPath: context.Path,
		Path:       context.Request.URL.Path,
		RawQuery:   context.Request.URL.RawQuery,
		ClientIP:   context.ClientIP,
		Method:     context.Request.Method,
		StatusCode: context.Response.StatusCode,
		//Header:r.Header
	}
	if t := context.Response.Header["Content-Type"]; len(t) > 0 {
		log.ContentType = t[0]
	}

	log.Latency = time.Now().Sub(context.StartTime)
	l := fmt.Sprintf("[QGate] %v  | %v | %v | %v | %v | %v | %v | %v Error:%v\n",
		log.ContentType, log.StatusCode, log.Latency, log.ClientIP, log.Method, log.Host, log.OriginPath, log.Path, log.Error)
	fmt.Fprintf(os.Stdout, l)
	qlog.Trace("access", l)
}
