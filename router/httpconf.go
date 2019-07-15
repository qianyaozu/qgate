package router

//import "sync"

type httpConf struct {
	Include      string        `json:"include"`
	Default_Type string        `json:"default_type"`
	Server       []*serverConf `json:"server"`
	Upstream     []*upstream   `json:"upstream"`
}

type upstream struct {
	Name    string   `json:"name"`
	Ip_Hash bool     `json:"ip_hash"`
	Server  []string `json:"server"`
	//Lock sync.Mutex
}

type location struct {
	//Match_Type int    `json:"match_type"`
	Regular    string `json:"regular"` //1:精准匹配，2:前缀优先匹配，3:正则匹配，4:前缀匹配
	Proxy_Pass string `json:"proxy_pass"`
	//proxy_redirect string
	//try_files
	//index
	//rewrite
	//root
	//deny
}

type serverConf struct {
	Default_Server bool        `json:"default_server"`
	Listen         int         `json:"listen"`
	Server_Name    string      `json:"server_name"` //优先级1:精准匹配，2:通配符在前匹配，3:通配符在后匹配，4:正则匹配
	Root           string      `json:"root"`
	Index          []string    `json:"index"`
	Location       []*location `json:"location"`
	Return         int         `json:"return"`
}
