package server

import (
	"fmt"
	"github.com/qianyaozu/qgate/handler"
	"github.com/qianyaozu/qgate/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start() {
	if err := router.LoadNginxConf("qgate.conf"); err != nil {
		fmt.Println("load config file error:", err)
		return
	}

	//获取监听端口列表
	ports := router.Conf.GetListenPorts()

	proxy := NewProxyServer()
	//注入路由转发中间件
	proxy.UseRequestHandler(&handler.QRouter{})

	//监听端口启动服务
	for _, p := range ports {
		go http.ListenAndServe(fmt.Sprintf(":%v", p), proxy)
	}

	//监听退出命令
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	for range signalChan {
		os.Exit(0)
	}
}
