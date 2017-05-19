package http

import (
	"net"
	"net/http"
	"strings"
)

// returns the client real ip address.
// inspired by https://github.com/gin-gonic/gin/blob/32cab500ecc71d2975f5699c8a65c6debb29cfbe/context.go#L341
func clientIP(r *http.Request) string {
	clientIP := strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}

	clientIP = r.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}

	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
