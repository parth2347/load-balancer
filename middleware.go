package load_balancer

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func WithLogging(h http.Handler) http.Handler {
	logFn := func(rw http.ResponseWriter, req * http.Request) {

		startTime := time.Now()
		uri := req.RequestURI
		method := req.Method

		h.ServeHTTP(rw, req)

		duration := time.Since(startTime)
		fmt.Println(duration)
		log.WithFields(log.Fields{
			"uri": uri,
			"method": method,
			"duration": duration,
		}).Info("request")
	}

	return http.HandlerFunc(logFn)
}
