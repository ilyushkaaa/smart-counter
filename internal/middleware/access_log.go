package middleware

import (
	"log"
	"net/http"
	"time"
)

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, err := GetLoggerFromContext(r.Context())
		if err != nil {
			log.Printf("can not get logger from context: %s", err)
			WriteNoLoggerResponse(w)
		}
		logger.Infof("access log middleware start")
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Infow("New request",
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
			"url", r.URL.Path,
			"time", time.Since(start),
		)
	})
}
