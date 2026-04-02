package result

import (
	"net/http"

	"github.com/gozero-hub/common/xerr"
)

func JwtUnauthorizedCallback(w http.ResponseWriter, r *http.Request, err error) {
	HttpResult(r, w, nil, nil, xerr.NewErrCode(xerr.UnauthorizedError))
}

func UnsignedCallback(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) {
	HttpResult(r, w, nil, nil, xerr.NewErrCode(xerr.UnauthorizedError))
}

func NotFoundCallback(w http.ResponseWriter, r *http.Request) {
	HttpResult(r, w, nil, nil, xerr.NewErrCode(xerr.NotFoundError))
}
