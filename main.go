package main

import (
	"fmt"
	"github.com/qianyaozu/qgate/handler"
	"github.com/qianyaozu/qgate/router"
	"github.com/qianyaozu/qgate/server"
	"github.com/qianyaozu/qlog"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//设置日志
	handler.Qlog = qlog.NewQLog(&qlog.Options{
		Name:         "qgate",
		RedisAddress: "192.168.2.207:6379",
		ChannelSize:  10000,
	})

	//读取配置
	if err := router.LoadNginxConf("qgateconfig.json"); err != nil {
		fmt.Println("load config file error:", err)
		return
	}

	//启动服务
	server.Start()
}
