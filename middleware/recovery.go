package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bryanmorgan/time-tracking-api/api"
)

func PanicHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				api.ErrorJson(w, api.NewError(fmt.Sprintf("%+v", err), r.RequestURI, api.PANIC), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
