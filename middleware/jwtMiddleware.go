package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gozero-hub/common/tool"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

func JwtMiddleware(jwtKey string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1️⃣ 检查 ctx 是否已有 UID（默认 JWT 中间件已经解析）
			if uid, ok := ctx.Value("Uid").(string); ok && uid != "" {
				ctx = logx.ContextWithFields(ctx, logx.LogField{Key: "login_uid", Value: uid})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2️⃣ 如果没有 UID，尝试手动解析 Authorization
			auth := r.Header.Get("Authorization")
			if auth != "" {
				parts := strings.SplitN(auth, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					token := parts[1]
					fmt.Println("Extracted Token:", token)
					claims, err := tool.ParseToken(token, jwtKey) // 使用传入的 jwtKey
					if err == nil && claims != nil {
						if uid, ok := (*claims)["Uid"].(string); ok && uid != "" {
							ctx = context.WithValue(ctx, "Uid", uid)
							ctx = logx.ContextWithFields(ctx, logx.LogField{Key: "login_uid", Value: uid})
							r = r.WithContext(ctx)
						}
					} else {
						logc.Errorf(ctx, "Token parsing error: %+v", err)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
