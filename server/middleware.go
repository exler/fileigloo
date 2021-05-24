package server

import (
	"log"
	"net/http"

	colors "github.com/logrusorgru/aurora"
)

func (s *Server) limitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.maxRequests == 0 {
			next.ServeHTTP(w, r)
			return
		}

		ip, err := getRealIP(r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		limiter := s.getVisitor(ip, s.maxRequests)
		if !limiter.Allow() {
			log.Printf("Rate limited IP: %s", colors.Red(ip))
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
