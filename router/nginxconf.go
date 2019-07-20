package router

import (
	"encoding/json"
	"errors"
	"github.com/qianyaozu/qgate/httphelper"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type QGateUser struct {
	UserName string
	Password string
	Token    string
}
type NginxConf struct {
	//StreamConf *streamConf `json:"stream"` //tcp配置
	HttpConf *httpConf    `json:"http"` //http配置
	User     []*QGateUser `json:"user"`
}

var Conf *NginxConf

//加载nginx配置文件
func LoadNginxConf(path string) error {
	var conf NginxConf
	contents, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(contents, &conf)
		if err == nil {
			Conf = &conf
			return nil
		}
	}
	return err
}

//获取需要监听的端口号
func (conf *NginxConf) GetListenPorts() []int {
	var contain = func(list []int, value int) bool {
		for _, item := range list {
			if item == value {
				return true
			}
		}
		return false
	}
	var ports = make([]int, 0)
	if len(conf.HttpConf.Server) > 0 {
		for _, s := range conf.HttpConf.Server {
			if s.Listen > 0 && !contain(ports, s.Listen) {
				ports = append(ports, s.Listen)
			}
		}
	}
	return ports
}

//获取配置服务
func (conf *NginxConf) GetHttpServer(r *http.Request) (server *ServerConf, err error) {
	domain := ""
	port := 0
	if strings.Contains(r.Host, ":") {
		hosts := strings.Split(r.Host, ":")
		domain = hosts[0]
		port, err = strconv.Atoi(hosts[1])
		if err != nil {
			return
		}
	} else {
		domain = r.Host
		port = 80
	}

	//匹配config规则
	servers := conf.getServerByPort(port)
	if len(servers) == 0 {
		err = errors.New("router->GetHttpLocation():没有匹配的Server条件")
		return
	}
	//获取符合匹配条件的服务
	server = getServerByDomain(domain, servers)
	return
}

//获取重定向的访问地址,返回重定向的地址，http返回标识，异常信息
func (server *ServerConf) GetHttpLocation(r *http.Request) (u *url.URL, returnCode int, err error) {
	u = nil
	returnCode = 0
	err = nil

	loc := getServerByName(r.URL.Path, server)
	if loc.Regular == "" {
		err = errors.New("router->GetHttpLocation():没有匹配的Location条件")
		return
	}
	u, err = getPathByLocation(httphelper.ClientIP(r), r.URL.Path, loc)
	if err != nil {
		return
	}
	return u, 0, nil
}

//根据端口获取服务配置列表
func (conf *NginxConf) getServerByPort(port int) []*ServerConf {
	var serverList = make([]*ServerConf, 0)
	for _, server := range conf.HttpConf.Server {
		if server.Listen == port {
			serverList = append(serverList, server)
		}
	}
	return serverList
}

//根据服务名称获取服务配置列表
func getServerByDomain(domain string, serverList []*ServerConf) *ServerConf {
	//匹配顺序：1:精准匹配，2:通配符在前匹配，3:通配符在后匹配，4:正则匹配 ，5:default_server为true,6:排在第一的
	matchServer := make([]*ServerConf, 0)
	for _, server := range serverList {
		if server.Server_Name == domain {
			//精准匹配
			return server
		} else if strings.HasPrefix(server.Server_Name, "*") && strings.HasSuffix(domain, strings.Replace(server.Server_Name, "*", "", 1)) {
			//通配符在前匹配
			matchServer = append(matchServer, server)
		} else if strings.HasSuffix(server.Server_Name, "*") && strings.HasPrefix(domain, strings.Replace(server.Server_Name, "*", "", 1)) {
			//通配符在后匹配
			matchServer = append(matchServer, server)
		} else {
			reg := regexp.MustCompile(server.Server_Name)
			if reg.MatchString(domain) {
				matchServer = append(matchServer, server)
			}
		}
	}
	for _, server := range matchServer {
		if strings.HasPrefix(server.Server_Name, "*") && strings.HasSuffix(domain, strings.Replace(server.Server_Name, "*", "", 1)) {
			return server
		}
	}
	for _, server := range matchServer {
		if strings.HasSuffix(server.Server_Name, "*") && strings.HasPrefix(domain, strings.Replace(server.Server_Name, "*", "", 1)) {
			return server
		}
	}
	for _, server := range matchServer {
		if reg := regexp.MustCompile(server.Server_Name); reg.MatchString(domain) {
			return server
		}
	}
	for _, server := range matchServer {
		if server.Default_Server {
			return server
		}
	}
	if len(matchServer) > 0 {
		return matchServer[0]
	}
	return serverList[0]
}

//根据服务获取location
func getServerByName(path string, server *ServerConf) *location {
	locs := make([]*location, 0)
	for _, loc := range server.Location {
		////1:精准匹配，2:前缀优先匹配，3:正则匹配，4:前缀匹配
		if strings.HasPrefix(loc.Regular, "=") && strings.Replace(loc.Regular, "=", "", 1) == path {
			return loc
		} else if strings.HasPrefix(loc.Regular, "^") && strings.HasPrefix(path, strings.Replace(loc.Regular, "^", "", 1)) {
			locs = append(locs, loc)
		} else if reg := regexp.MustCompile(loc.Regular); reg.MatchString(path) {
			locs = append(locs, loc)
		} else if strings.HasPrefix(path, loc.Regular) {
			locs = append(locs, loc)
		}
	}
	locmap := make(map[int]*location)
	//判断前缀优先匹配
	for _, loc := range locs {
		if strings.HasPrefix(loc.Regular, "^") && strings.HasPrefix(path, strings.Replace(loc.Regular, "^", "", 1)) {
			if l, ok := locmap[2]; ok {
				if len(loc.Regular) > len(l.Regular) {
					locmap[2] = loc
				}
			} else {
				locmap[2] = loc
			}
		}
	}
	if l, ok := locmap[2]; ok {
		return l
	}
	//判断正则匹配
	for _, loc := range locs {
		if reg := regexp.MustCompile(loc.Regular); reg.MatchString(path) {
			return loc
		}
	}
	//判断前缀优先匹配
	for _, loc := range locs {
		if strings.HasPrefix(path, loc.Regular) {
			if l, ok := locmap[4]; ok {
				if len(loc.Regular) > len(l.Regular) {
					locmap[4] = loc
				}
			} else {
				locmap[4] = loc
			}
		}
	}
	if l, ok := locmap[4]; ok {
		return l
	}

	return &location{Regular: ""}
}

//根据路径或路由条件获取实际工作路径
func getPathByLocation(ip, path string, loc *location) (*url.URL, error) {
	//是否是相对路径
	var relative = true
	if strings.HasSuffix(loc.Proxy_Pass, "/") {
		relative = false
	}
	u, err := url.Parse(loc.Proxy_Pass)
	if err != nil {
		return nil, err
	}
	//根据规则替换path
	if relative {
		u.Path = u.Path + path
	} else {
		matchPath := ""
		if reg := regexp.MustCompile(loc.Regular); reg.MatchString(path) {
			matchPath = reg.FindString(path)
		} else {
			matchPath = loc.Regular
		}
		u.Path = u.Path + strings.Replace(path, matchPath, "", 1)
	}
	//根据规则替换upstream
	for _, upstream := range Conf.HttpConf.Upstream {
		if upstream.Name == u.Host {
			if upstream.Ip_Hash {
				//iphash 获取服务器IP
				ips := strings.Split(ip, ".")
				hash := 0
				for _, i := range ips {
					value, err := strconv.Atoi(i)
					if err == nil {
						hash += value
					}
				}
				u.Host = upstream.Server[hash%len(upstream.Server)]
			} else {
				//随机获取ip
				//todo 增加weight和检测可用性
				u.Host = upstream.Server[rand.Intn(len(upstream.Server))]
			}
			break
		}
	}
	return u, nil
}
