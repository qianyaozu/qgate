package handler

import (
	"encoding/base64"
	"github.com/qianyaozu/qgate/httphelper"
	"github.com/qianyaozu/qgate/router"
	"strings"
)

type QAuth struct {
}

//身份认证
func (qauth *QAuth) Handle(context *QContext) {
	switch strings.ToLower(context.Server.Auth) {
	case "":
	case "none":
		{
			break
		}
	case "token":
	case "basic":
		{
			token := context.Request.Header["Authorization"]
			if len(token) == 0 {
				context.Response = httphelper.NewResponse(httphelper.UNAUTHORIZED, "UNAUTHORIZED")
				return
			}

			for _, user := range router.Conf.User {
				if strings.ToLower(context.Server.Auth) == "token" {
					//token 认证
					if user.Token == token[0] {
						context.UserName = user.UserName
						return
					}
				} else {
					//basic认证
					encodeString := base64.StdEncoding.EncodeToString([]byte(user.UserName + ":" + user.Password))
					if encodeString == token[0] {
						context.UserName = user.UserName
						return
					}
				}
			}
			context.Response = httphelper.NewResponse(httphelper.UNAUTHORIZED, "UNAUTHORIZED")
			return

		}
	}
	return
}
