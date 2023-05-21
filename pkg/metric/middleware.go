package metric

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func (m *Metric) GetPrometheusMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()

			timer := prometheus.NewTimer(m.httpDuration.WithLabelValues(path))
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			statusCode := rw.statusCode

			m.responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
			m.totalRequests.WithLabelValues(path).Inc()

			timer.ObserveDuration()
		})
	}
}
