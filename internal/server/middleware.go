package server

import "net/http"

func LogMiddleware(next http.Handler, l Logger) http.Handler {
	l.Info("[ + ] Middleware was set up")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Fatal(err)
			}
		}()

		info := NewLogInfo(r)
		next.ServeHTTP(w, r)
		l.Info(info.String())
	})
}
