package middleware

import (
	"net/http"
)

func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Request-ID, Accept, X-CSRF-Token, X-Request-Id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("X-Frame-Options", "ALLOWALL")

		// 处理预检请求
		if r.Method == http.MethodOptions {
			// 对于 OPTIONS 请求，直接返回 200 状态码
			w.WriteHeader(http.StatusNoContent)
			//httpx.WriteJson(w, http.StatusOK, map[string]interface{}{})
			return
		}

		next(w, r)
	}
}
