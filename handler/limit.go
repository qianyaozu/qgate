package handler

import (
	"encoding/base64"
	"github.com/qianyaozu/qgate/httphelper"
	"strings"
	"sync"
	"time"
)

type QLimit struct {
}

const (
	Period_Second string = "Second"
	Period_Minute string = "Minute"
	Period_Hour   string = "Hour"
	Period_Day    string = "Day"
	Period_Week   string = "Week"
)

//时间戳字典
var timePeriodMap = make(map[string]int64)

func init() {
	timePeriodMap["Second"] = 1 * 1000
	timePeriodMap["Minute"] = 1 * 60 * 1000
	timePeriodMap["Hour"] = 1 * 60 * 60 * 1000
	timePeriodMap["Day"] = 1 * 60 * 60 * 24 * 1000
	timePeriodMap["Week"] = 1 * 60 * 60 * 24 * 7 * 1000
}

//限流
func (qlimit *QLimit) Handle(context *QContext) {
	if context.IsInWhiteList || len(context.Server.Limit) == 0 {
		//IP在白名单中或者没有设置限流规则，则不进行校验
		return
	}

	for _, limit := range context.Server.Limit {
		if len(limit.Policy) == 0 {
			limit.Policy = []string{"ip", "path"}
		}
		// region 循环判断 每秒,每分钟,每小时,每天,每周的限流情况
		if limit.Second > 0 && !qlimit.allowRequest(Period_Second, limit.Policy, limit.Second, context) {
			context.Response = httphelper.NewResponse(httphelper.TOO_MANY_REQUESTS, "TOO_MANY_REQUESTS_PER_SECOND")
			return
		}
		if limit.Minute > 0 && !qlimit.allowRequest(Period_Minute, limit.Policy, limit.Minute, context) {
			context.Response = httphelper.NewResponse(httphelper.TOO_MANY_REQUESTS, "TOO_MANY_REQUESTS_PER_MINUTE")
			return
		}
		if limit.Hour > 0 && !qlimit.allowRequest(Period_Hour, limit.Policy, limit.Hour, context) {
			context.Response = httphelper.NewResponse(httphelper.TOO_MANY_REQUESTS, "TOO_MANY_REQUESTS_PER_HOUR")
			return
		}
		if limit.Day > 0 && !qlimit.allowRequest(Period_Day, limit.Policy, limit.Day, context) {
			context.Response = httphelper.NewResponse(httphelper.TOO_MANY_REQUESTS, "TOO_MANY_REQUESTS_PER_DAY")
			return
		}
		if limit.Week > 0 && !qlimit.allowRequest(Period_Week, limit.Policy, limit.Week, context) {
			context.Response = httphelper.NewResponse(httphelper.TOO_MANY_REQUESTS, "TOO_MANY_REQUESTS_PER_WEEK")
			return
		}
		// endregion
	}

}

//根据策略获取请求标识
func (qlimit *QLimit) getRequestIdentityToken(context *QContext, period string, policy []string) string {
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

var throttlingMap sync.Map     //限流字典
var throttlingMutex sync.Mutex //限流锁

//根据唯一标识获取限流计数器
func (qlimit *QLimit) allowRequest(period string, policy []string, limit int, context *QContext) bool {
	//根据策略获取请求标识
	identityToken := qlimit.getRequestIdentityToken(context, period, policy)

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
	if (c.Timestamp + timePeriodMap[period]) < time.Now().Unix() {
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
