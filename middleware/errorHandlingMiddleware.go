package middleware

import (
	"net/http"

	"github.com/gozero-hub/common/result"

	"github.com/pkg/errors"
)

func ErrorHandlingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); any(err) != nil {
				// fmt.Printf("aaaaaaaaaaaaaaaaaaaaaaaaa%+v",err)
				// http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				result.HttpResult(r, w, nil, nil, errors.Errorf("%+v", err))
			}
		}()
		next.ServeHTTP(w, r)
	}
}
