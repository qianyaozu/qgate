package httphelper

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) string {
	if clientIPs := r.Header["X-Forwarded-For"]; len(clientIPs) > 0 {
		clientIP := clientIPs[0]
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if len(clientIP) > 0 {
			return clientIP
		}

	}
	if clientIPs := r.Header["X-Real-Ip"]; len(clientIPs) > 0 {
		return clientIPs[0]
	}
	if clientIPs := r.Header["X-Appengine-Remote-Addr"]; len(clientIPs) > 0 {
		return clientIPs[0]
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
