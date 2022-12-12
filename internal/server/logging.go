package server

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type LogInfo struct {
	date        string
	latency     string
	ip          string
	httpMethod  string
	path        string
	httpVersion string
	respStatus  string
	userAgent   string
}

func NewLogInfo(r *http.Request) *LogInfo {
	latency := time.Now()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	info := &LogInfo{
		ip:          ip,
		date:        time.Now().UTC().Format("2006-01-02 15:04:05"),
		httpVersion: r.Proto,
		httpMethod:  r.Method,
		path:        r.URL.Path,
		respStatus:  "mb 200", // r.Response.Status,
		userAgent:   r.UserAgent(),
	}

	if r.Response != nil {
		info.respStatus = r.Response.Status
	}

	info.latency = time.Since(latency).String()

	return info
}

func (i *LogInfo) String() string {
	return fmt.Sprintf("%s [%s] %s %s %s %q %s %s",
		i.ip,
		i.date,
		i.httpMethod,
		i.path,
		i.httpVersion,
		i.respStatus,
		i.latency,
		i.userAgent,
	)
}
