package server

import (
	"net/http"
)

func logMiddleware(next http.Handler) http.Handler {
	l.Info("[ + ] Middleware Started")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Fatal(err)
			}
		}()

		info := NewHTTPInfo(r)
		next.ServeHTTP(w, r)
		l.Info(info.String())
	})
}
