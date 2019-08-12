package handler

import (
	"encoding/base64"
	"fmt"
	"github.com/qianyaozu/qgate/httphelper"
	"net/http"
	"strings"
	"sync"
	"time"
)

type QLimit struct {
}

const (
	Period_Second = "Second"
	Period_Minute = "Minute"
	Period_Hour   = "Hour"
	Period_Day    = "Day"
	Period_Week   = "Week"
)

//时间戳字典
var timePeriodMap = map[string]int64{
	Period_Second: 1,
	Period_Minute: 1 * 60,
	Period_Hour:   1 * 60 * 60,
	Period_Day:    1 * 60 * 60 * 24,
	Period_Week:   1 * 60 * 60 * 24 * 7,
}

//限流
func (qlimit *QLimit) Handle(context *QContext) {
	var token = getServerRequestIdentityToken(context)
	if con, ok := ConnectionLimitMap.Load(token); ok {
		c := con.(*ConnectLimit)
		c.Mutex.Lock()
		c.CurrentConnection = c.CurrentConnection + 1
		c.Mutex.Unlock()
	} else {
		ConnectionLimitMap.Store(token, &ConnectLimit{CurrentConnection: 1})
	}

	if context.IsInWhiteList || context.Server == nil || len(context.Server.Limit) == 0 {
		//IP在白名单中或者没有设置限流规则，则不进行校验
		return
	}

	for _, limit := range context.Server.Limit {
		if len(limit.Policy) == 0 {
			//如果没有设置策略，则默认以IP 和访问路径作为 检测策略
			limit.Policy = []string{"ip", "path"}
		}

		//并发连接数限制
		if limit.Connection > 0 {
			var token = getRequestIdentityToken(context, "", limit.Policy)
			if con, ok := ConnectionLimitMap.Load(token); ok {
				c := con.(*ConnectLimit)
				if c.CurrentConnection >= limit.Connection {
					context.Response = httphelper.NewResponse(http.StatusTooManyRequests, "TOO_MANY_REQUESTS_CONCURRENCY")
					return
				} else {
					if !context.IsAddConnectionCount {
						context.IsAddConnectionCount = true
						c.Mutex.Lock()
						c.CurrentConnection = c.CurrentConnection + 1
						c.Mutex.Unlock()

					}
				}
			} else {
				ConnectionLimitMap.Store(token, &ConnectLimit{CurrentConnection: 1})
				context.IsAddConnectionCount = true
			}
		}

		// region 循环判断 每秒,每分钟,每小时,每天,每周的限流情况
		if limit.Second > 0 && !qlimit.allowRequest(Period_Second, limit.Policy, limit.Second, context) {
			context.Response = httphelper.NewResponse(http.StatusTooManyRequests, "TOO_MANY_REQUESTS_PER_SECOND")
			return
		}
		if limit.Minute > 0 && !qlimit.allowRequest(Period_Minute, limit.Policy, limit.Minute, context) {
			context.Response = httphelper.NewResponse(http.StatusTooManyRequests, "TOO_MANY_REQUESTS_PER_MINUTE")
			return
		}
		if limit.Hour > 0 && !qlimit.allowRequest(Period_Hour, limit.Policy, limit.Hour, context) {
			context.Response = httphelper.NewResponse(http.StatusTooManyRequests, "TOO_MANY_REQUESTS_PER_HOUR")
			return
		}
		if limit.Day > 0 && !qlimit.allowRequest(Period_Day, limit.Policy, limit.Day, context) {
			context.Response = httphelper.NewResponse(http.StatusTooManyRequests, "TOO_MANY_REQUESTS_PER_DAY")
			return
		}
		if limit.Week > 0 && !qlimit.allowRequest(Period_Week, limit.Policy, limit.Week, context) {
			context.Response = httphelper.NewResponse(http.StatusTooManyRequests, "TOO_MANY_REQUESTS_PER_WEEK")
			return
		}
		// endregion
	}

}

//根据策略获取请求标识
func getRequestIdentityToken(context *QContext, period string, policy []string) string {
	identityToken := period
	for _, p := range policy {
		switch strings.ToLower(p) {
		case "auth":
			{
				identityToken += "_" + context.UserName
				break
			}
		case "ip":
			{
				identityToken += "_" + context.ClientIP
				break
			}
		case "path":
			{
				identityToken += "_" + context.Path
				break
			}
		case "agent":
			{
				identityToken += "_" + context.Request.UserAgent()
				break
			}
		}
	}
	return base64.StdEncoding.EncodeToString([]byte(identityToken))
}

//获取服务对应的token
func getServerRequestIdentityToken(context *QContext) string {
	var name = "server"
	if context.Server != nil {
		name = context.Server.Server_Name + fmt.Sprint(context.Server.Listen)
	}
	return base64.StdEncoding.EncodeToString([]byte(name))
}

var throttlingMap sync.Map     //限流字典
var throttlingMutex sync.Mutex //限流锁

//根据唯一标识获取限流计数器
func (qlimit *QLimit) allowRequest(period string, policy []string, limit int, context *QContext) bool {
	//根据策略获取请求标识
	identityToken := getRequestIdentityToken(context, period, policy)
	throttlingMutex.Lock()
	defer throttlingMutex.Unlock()

	newCounter := &ThrottlingCounter{
		Timestamp:    time.Now().Unix(),
		RequestCount: 1,
	}

	counter, isload := throttlingMap.LoadOrStore(identityToken, newCounter)
	if !isload {
		context.IsAddRequestCount = true
	}

	c := counter.(*ThrottlingCounter)
	//在单位时间内，则计数增加，否则重置时间戳和计数
	if (c.Timestamp + timePeriodMap[period]) > time.Now().Unix() {
		if !context.IsAddRequestCount {
			c.RequestCount += 1
			context.IsAddRequestCount = true
		}
	} else {
		if !context.IsAddRequestCount {
			c.RequestCount = 1
			c.Timestamp = newCounter.Timestamp
			context.IsAddRequestCount = true
		}
	}
	throttlingMap.Store(identityToken, c)

	if c.RequestCount > limit {
		return false
	}
	return true
}

//限流计数器
type ThrottlingCounter struct {
	Timestamp    int64 //时间戳
	RequestCount int   //单位时间内合计请求次数

}

//并发请求计数map
//var ConnectionLimitMap = make(map[string]*ConnectLimit)
var ConnectionLimitMap sync.Map //限流字典
type ConnectLimit struct {
	CurrentConnection int        //当前连接数
	Mutex             sync.Mutex //并发连接计数器锁
}

//释放并发连接计数
func ReleaseConnectionLimit(context *QContext) (current int) {
	current = 0
	var token = getServerRequestIdentityToken(context)

	if con, ok := ConnectionLimitMap.Load(token); ok {
		c := con.(*ConnectLimit)
		c.Mutex.Lock()
		c.CurrentConnection = c.CurrentConnection - 1
		current = c.CurrentConnection
		c.Mutex.Unlock()
	}
	if context.Server == nil || len(context.Server.Limit) == 0 {
		return
	}
	for _, limit := range context.Server.Limit {
		if len(limit.Policy) == 0 {
			//如果没有设置策略，则默认以IP 和访问路径作为 检测策略
			limit.Policy = []string{"ip", "path"}
		}

		//并发连接数限制
		if limit.Connection > 0 {
			var token = getRequestIdentityToken(context, "", limit.Policy)
			if con, ok := ConnectionLimitMap.Load(token); ok {
				c := con.(*ConnectLimit)
				if context.IsAddConnectionCount {
					context.IsAddConnectionCount = false
					c.Mutex.Lock()
					c.CurrentConnection = c.CurrentConnection - 1
					c.Mutex.Unlock()
				}
			}
		}
	}
	return
}
