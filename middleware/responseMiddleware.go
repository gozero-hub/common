package middleware

import (
	"fmt"
	"net/http"

	"gitee.com/scholar-hub/go-zero-common/result"
	"gitee.com/scholar-hub/go-zero-common/xerr"
)

func ResponseMiddlewarefunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w}
		next(rw, r)
		fmt.Println(rw)

		if rw.status == 404 {
			result.HttpResult(r, w, nil, nil, xerr.NewErrCode(xerr.NotFoundError))
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
