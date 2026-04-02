package result

import (
	"context"
	"github.com/zeromicro/go-zero/core/trace"
)

type ResponseBean struct {
	Code      uint32      `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}
type NullJson struct{}

func Success(ctx context.Context, data interface{}) *ResponseBean {
	return &ResponseBean{200, "OK", trace.TraceIDFromContext(ctx), data}
}

func Error(ctx context.Context, errCode uint32, errMsg string) *ResponseBean {
	return &ResponseBean{errCode, errMsg, trace.TraceIDFromContext(ctx), nil}
}
