package result

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gozero-hub/common/tool"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/gozero-hub/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
)

func HttpResult(r *http.Request, w http.ResponseWriter, req interface{}, resp interface{}, err error) {
	//fmt.Printf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%+v", err)
	if err == nil {
		//成功返回
		r := Success(r.Context(), resp)
		httpx.WriteJson(w, http.StatusOK, r)
	} else {
		//错误返回
		errCode := xerr.ServerCommonError
		errMsg := err.Error()

		causeErr := errors.Cause(err)                // err类型
		if e, ok := causeErr.(*xerr.CodeError); ok { //自定义错误类型
			errCode = e.GetErrCode()
			errMsg = e.GetErrMsg()
		} else {
			if gStatus, ok := status.FromError(causeErr); ok { // grpc err错误
				grpcCode := uint32(gStatus.Code())
				if xerr.IsCodeErr(grpcCode) { //区分自定义错误跟系统底层、db等错误，底层、db错误不能返回给前端
					errCode = grpcCode
					errMsg = gStatus.Message()
				}
			}
		}

		logc.Errorw(r.Context(), "request fail",
			logc.Field("url", tool.FullURL(r)),
			logc.Field("error_code", errCode),
			logc.Field("error_msg", errMsg),
			logc.Field("params", GetAllParams(r)),
			logc.Field("req", req),
			logc.Field("resp", resp),
			logc.Field("error_origin", fmt.Sprintf("%+v", err)),
		)

		responseCode := int(errCode)
		if responseCode > 500 {
			responseCode = 500
		}
		if responseCode == 500 {
			errMsg = "server internal error"
		}

		httpx.WriteJson(w, responseCode, Error(r.Context(), errCode, errMsg))
	}
}

func ParamErrorResult(r *http.Request, w http.ResponseWriter, err error) {
	httpx.WriteJson(w, http.StatusBadRequest, Error(r.Context(), xerr.RequestParamError, err.Error()))
}

// GetAllParams 把 r 中能拆的参数全拆出来
func GetAllParams(r *http.Request) map[string]any {
	out := make(map[string]any)

	// 1. Query 参数
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			out[k] = v[0] // 单值直接用 string
		} else {
			out[k] = v // 多值用 []string
		}
	}

	// 2. Form 参数（application/x-www-form-urlencoded 或 multipart）
	if err := r.ParseForm(); err == nil {
		for k, v := range r.Form {
			if len(v) == 1 {
				out[k] = v[0]
			} else {
				out[k] = v
			}
		}
	}

	// 4. JSON Body（只读一次，可再 io.TeeReader 留底）
	if r.Header.Get("Content-Type") == "application/json" {
		var body map[string]any
		// 把 body 拷一份出来，避免读空
		buf, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(buf, &body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(buf)) // 回填，后续还能用
		for k, v := range body {
			out[k] = v
		}
	}

	return out
}
