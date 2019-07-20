package handler

import (
	"github.com/qianyaozu/qgate/httphelper"
	"github.com/qianyaozu/qgate/router"
	"net/http"
	"time"
)

//定义事件处理接口
type QHandler interface {
	Handle(context *QContext)
}

type QContext struct {
	Request       *http.Request
	Response      *http.Response
	ClientIP      string
	IsInWhiteList bool //是否存在于白名单中
	Path          string
	StartTime     time.Time          //开始统计时间
	EndTime       time.Time          //结束统计时间
	Latency       time.Duration      //请求耗时
	Server        *router.ServerConf //路由信息
	UserName      string             //用户名

	IsAddRequestCount bool //是否已经增加请求计数
}

//初始化上下文
func InitContext(req *http.Request) *QContext {
	context := &QContext{
		Request:   req,
		ClientIP:  httphelper.ClientIP(req),
		StartTime: time.Now(),
		Path:      req.URL.Path,
	}
	return context
}
