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
	if err := router.LoadNginxConf("D:\\go\\src\\github.com\\qianyaozu\\qgate\\router\\qgate.conf"); err != nil {
		fmt.Println("load config file error:", err)
		return
	}

	ports := router.Conf.GetListenPorts() //获取监听端口列表

	proxy := NewProxyServer()
	proxy.UseRequestHandler(&handler.QRouter{})
	for _, p := range ports {
		go http.ListenAndServe(fmt.Sprintf(":%v", p), proxy)
	}
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
