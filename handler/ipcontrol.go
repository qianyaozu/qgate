package handler

import (
	"github.com/qianyaozu/qgate/httphelper"
	"net/http"
)

type QIPControl struct {
}

//IP 白名单  黑名单 控制
func (control *QIPControl) Handle(context *QContext) {
	if len(context.Server.Allow) > 0 {
		//白名单
		for _, allow := range context.Server.Allow {
			if context.ClientIP == allow {
				context.IsInWhiteList = true
				return
			}
		}
		context.Response = httphelper.NewResponse(http.StatusBadGateway, "IP FORBIDDEN")
	} else if len(context.Server.Deny) > 0 {
		//黑名单
		for _, deny := range context.Server.Deny {
			if context.ClientIP == deny {
				context.Response = httphelper.NewResponse(http.StatusBadGateway, "IP FORBIDDEN")
			}
		}
	}
	return
}
