package middleware

import (
	"log"
	"net/http"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()

			if rec != nil {
				log.Printf("panic: %v", rec)
				w.WriteHeader(http.StatusInternalServerError)

			}

		}()

		next.ServeHTTP(w, r)
	})
}
