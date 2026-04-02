package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

func IpMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		ctx := context.WithValue(r.Context(), "ip", ip)
		// ip = tool.GetIPFromCtx(ctx)
		// 3. 绑定到 logx
		ctx = logx.ContextWithFields(ctx, logx.LogField{Key: "clinet_ip", Value: ip})

		next(w, r.WithContext(ctx))
	}
}

// getClientIP 从 http.Request 中获取客户端IP
func getClientIP(r *http.Request) string {
	// X-Forwarded-For 可能包含多个 IP，第一个为真实IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
