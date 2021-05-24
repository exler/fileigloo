package server

import (
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (s *Server) getVisitor(ip string, maxRequests int) *rate.Limiter {
	s.visitorsMutex.Lock()
	defer s.visitorsMutex.Unlock()

	v, exists := s.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(maxRequests), 3)
		s.visitors[ip] = &Visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (s *Server) cleanVisitors() {
	s.visitorsMutex.Lock()
	defer s.visitorsMutex.Unlock()

	for ip, v := range s.visitors {
		if time.Since(v.lastSeen) > 3*time.Minute {
			delete(s.visitors, ip)
		}
	}
}

func getRealIP(r *http.Request) (ip string, err error) {
	var addr string

	if r.Header.Get("X-Forwarded-For") != "" {
		addr = r.Header.Get("X-Forwarded-For")
	} else if r.Header.Get("X-Real-IP") != "" {
		addr = r.Header.Get("X-Real-IP")
	} else {
		addr = r.RemoteAddr
	}

	ip, _, err = net.SplitHostPort(addr)
	return
}
