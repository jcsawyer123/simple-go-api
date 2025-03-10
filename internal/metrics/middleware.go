package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// HTTPMiddleware creates a middleware that records HTTP request metrics
func HTTPMiddleware(metricsPrefix string) func(next http.Handler) http.Handler {
	// Pre-create metrics with vectored labels
	requestsTotal := CounterMetric(metricsPrefix+"_requests_total", map[string]string{
		"method": "",
		"route":  "",
		"status": "",
	})

	requestDuration := HistogramMetric(metricsPrefix+"_request_duration_seconds", map[string]string{
		"method": "",
		"route":  "",
		"status": "",
	})

	// Simple in-flight counter doesn't need labels
	requestsInFlight := GaugeMetric(metricsPrefix+"_requests_in_flight", nil)

	requestSize := HistogramMetric(metricsPrefix+"_request_size_bytes", map[string]string{
		"method": "",
		"route":  "",
	})

	responseSize := HistogramMetric(metricsPrefix+"_response_size_bytes", map[string]string{
		"method": "",
		"route":  "",
		"status": "",
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract route pattern if using chi router
			route := "unknown"
			if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
				if routePattern := routeCtx.RoutePattern(); routePattern != "" {
					route = routePattern
				}
			}

			// Increment in-flight requests counter
			requestsInFlight.Inc()
			defer requestsInFlight.Dec()

			// Track request size if Content-Length is set
			if r.ContentLength > 0 {
				requestSize.With(map[string]string{
					"method": r.Method,
					"route":  route,
				}).Observe(float64(r.ContentLength))
			}

			// Use a response writer wrapper to capture status code and size
			ww := newResponseWriter(w)

			// Process the request
			next.ServeHTTP(ww, r)

			// Record metrics after request is completed
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.status)

			// Record request count and duration with appropriate labels
			labels := map[string]string{
				"method": r.Method,
				"route":  route,
				"status": status,
			}

			requestsTotal.With(labels).Inc()
			requestDuration.With(labels).Observe(duration)

			// Record response size
			if ww.written > 0 {
				responseSize.With(labels).Observe(float64(ww.written))
			}
		})
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures status code and response size
type responseWriter struct {
	http.ResponseWriter
	status  int
	written int64
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK, // Default status
	}
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Unwrap returns the underlying ResponseWriter
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
