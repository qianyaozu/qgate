package handler

import (
	"fmt"
	"github.com/qianyaozu/qlog"
	"strings"
	"time"
)

var Qlog *qlog.QLog

type QRequestLog struct {
	Name         string `json:"name"` //用于elk存储
	Latency      int64  //请求耗时
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
	Connection  int //并发请求数量
	Error       interface{}
}

func HandleLog(context *QContext) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("HandleLog", err)
		}
	}()
	var log = QRequestLog{
		Name:       Qlog.Name + "_" + context.Server.Server_Name,
		Host:       context.Request.Host,
		OriginPath: context.Path,
		Path:       context.Request.URL.Path,
		RawQuery:   context.Request.URL.RawQuery,
		ClientIP:   context.ClientIP,
		Method:     context.Request.Method,
		StatusCode: context.Response.StatusCode,
		Connection: context.RemainConnection,
		Error:      context.Error,
	}
	if t := context.Response.Header["Content-Type"]; len(t) > 0 {
		log.ContentType = t[0]
	}

	log.Latency = time.Now().Sub(context.StartTime).Nanoseconds() / 1000000
	l := fmt.Sprintf("[QGate] %v  | %v | %v | %v | %v | %v | %v | %v Connection:%v   Error:%v",
		log.ContentType, log.StatusCode, log.Latency, log.ClientIP, log.Method, log.Host, log.OriginPath, log.Path, log.Connection, log.Error)
	fmt.Println(l)
	Qlog.Trace("access", l)
	if strings.HasPrefix(log.ContentType, "application/json") {
		Qlog.Elk("access", log)
	}
}
