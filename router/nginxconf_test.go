package router

import (
	"encoding/json"
	"fmt"
	"testing"
)

func BenchmarkFirst(b *testing.B) {
	conf, err := LoadNginxConf("D:\\go\\src\\github.com\\qianyaozu\\qgate\\router\\qgate.conf")
	//fmt.Println(conf)
	js, err := json.Marshal(conf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(js))

	conf.GetHttpLocation("192.168.2.162", "localhost:8080", "/image/aa")
}

func Benchmark1(b *testing.B) {
	conf, _ := LoadNginxConf("D:\\go\\src\\github.com\\qianyaozu\\qgate\\router\\qgate.conf")
	//fmt.Println(conf)

	fmt.Println(conf.GetListenPorts())
}
