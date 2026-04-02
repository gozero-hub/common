package {{.PkgName}}

import (
    "errors"
    "net/http"

    "github.com/gozero-hub/common/result"
    {{if .HasRequest}}"github.com/gozero-hub/common/xerr"{{end}}
    {{if .HasRequest}}"github.com/gozero-hub/common/translator"{{end}}
    {{if .HasRequest}}"github.com/zeromicro/go-zero/rest/httpx"{{end}}
    {{.ImportPackages}}
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        {{if .HasRequest}}var req types.{{.RequestType}}
        if err := httpx.Parse(r, &req); err != nil {
            result.ParamErrorResult(r, w, errors.New(xerr.MapErrMsg(xerr.RequestParamError)))
            return
        }

        validateErr := translator.ValidateStruct(&req)
        if validateErr != nil {
            result.ParamErrorResult(r, w, validateErr)
            return
        }

        {{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
        {{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})

        result.HttpResult(r, w, {{if .HasRequest}}req{{else}}nil{{end}}, {{if .HasResp}}resp{{else}}nil{{end}}, err)
    }
}