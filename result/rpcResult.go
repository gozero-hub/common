package result

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logc"

	"github.com/gozero-hub/common/xerr"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
)

func RpcResult(ctx context.Context, req interface{}, resp interface{}, err error) *ResponseBean {
	if err == nil {
		return Success(ctx, resp)
	} else {
		// 错误返回
		errCode := xerr.ServerCommonError
		errMsg := err.Error()

		causeErr := errors.Cause(err)                // err类型
		if e, ok := causeErr.(*xerr.CodeError); ok { // 自定义错误类型
			errCode = e.GetErrCode()
			errMsg = e.GetErrMsg()
		} else {
			if gStatus, ok := status.FromError(causeErr); ok { // grpc err错误
				grpcCode := uint32(gStatus.Code())
				if xerr.IsCodeErr(grpcCode) { // 区分自定义错误跟系统底层、db等错误，底层、db错误不能返回给前端
					errCode = grpcCode
					errMsg = gStatus.Message()
				}
			}
		}

		logc.Errorw(ctx, "rpc request fail",
			logc.Field("error_code", errCode),
			logc.Field("error_msg", errMsg),
			logc.Field("req", req),
			logc.Field("resp", resp),
			logc.Field("error_origin", fmt.Sprintf("%+v", err)),
		)

		responseCode := int(errCode)
		if responseCode > 500 {
			responseCode = 500
		}

		return Error(ctx, errCode, errMsg)
	}
}
