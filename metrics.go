package main

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Minimal Prometheus metrics (text format, no dependency): requests by route × status code and a latency histogram.
// Used by the canary analysis (error rate / p95 of the canary pods) and the SLO alerts.

var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}

type metrics struct {
	mu       sync.Mutex
	requests map[[2]string]uint64 // {route, code}
	hist     map[string][]uint64  // route → per-bucket counts (+Inf last)
	sum      map[string]float64
}

func newMetrics() *metrics {
	return &metrics{requests: map[[2]string]uint64{}, hist: map[string][]uint64{}, sum: map[string]float64{}}
}

func (m *metrics) observe(route string, code int, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[[2]string{route, strconv.Itoa(code)}]++
	h := m.hist[route]
	if h == nil {
		h = make([]uint64, len(buckets)+1)
		m.hist[route] = h
	}
	i := sort.SearchFloat64s(buckets, d.Seconds())
	h[i]++
	m.sum[route] += d.Seconds()
}

func (m *metrics) write(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Fprintln(w, "# TYPE http_requests_total counter")
	keys := make([][2]string, 0, len(m.requests))
	for k := range m.requests {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i][0]+keys[i][1] < keys[j][0]+keys[j][1] })
	for _, k := range keys {
		fmt.Fprintf(w, "http_requests_total{route=%q,code=%q} %d\n", k[0], k[1], m.requests[k])
	}
	fmt.Fprintln(w, "# TYPE http_request_duration_seconds histogram")
	routes := make([]string, 0, len(m.hist))
	for r := range m.hist {
		routes = append(routes, r)
	}
	sort.Strings(routes)
	for _, r := range routes {
		var cum uint64
		for i, b := range buckets {
			cum += m.hist[r][i]
			fmt.Fprintf(w, "http_request_duration_seconds_bucket{route=%q,le=%q} %d\n", r, strconv.FormatFloat(b, 'g', -1, 64), cum)
		}
		cum += m.hist[r][len(buckets)]
		fmt.Fprintf(w, "http_request_duration_seconds_bucket{route=%q,le=\"+Inf\"} %d\n", r, cum)
		fmt.Fprintf(w, "http_request_duration_seconds_sum{route=%q} %g\n", r, m.sum[r])
		fmt.Fprintf(w, "http_request_duration_seconds_count{route=%q} %d\n", r, cum)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (s *statusRecorder) WriteHeader(c int) { s.code = c; s.ResponseWriter.WriteHeader(c) }

// instrument records every request under its route pattern (not the raw path: bounded cardinality).
func (m *metrics) instrument(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, code: 200}
		mux.ServeHTTP(rec, r)
		_, pattern := mux.Handler(r)
		if pattern == "" {
			pattern = "unmatched"
		}
		if pattern != "GET /metrics" {
			m.observe(pattern, rec.code, time.Since(start))
		}
	})
}
