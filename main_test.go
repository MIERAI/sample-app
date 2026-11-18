package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMux(t *testing.T) {
	h := newMux("abc123", false, newMetrics())
	for _, c := range []struct {
		path, wantBody string
		wantCode       int
	}{
		{"/healthz", "ok", 200},
		{"/hello", `"message":"hello, world"`, 200},
		{"/hello?name=gitops", `"message":"hello, gitops"`, 200},
		{"/hello?name=gitops&lang=ja", `"message":"こんにちは、gitops"`, 200},
		{"/hello?name=" + strings.Repeat("x", 65), "too long", 400},
		{"/nope", "", 404},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", c.path, nil))
		if w.Code != c.wantCode || !strings.Contains(w.Body.String(), c.wantBody) {
			t.Errorf("%s: %d %q", c.path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/healthz", nil))
	if v := w.Header().Get("X-App-Version"); v != "abc123" {
		t.Errorf("version header = %q", v)
	}
}

func TestFaultHello(t *testing.T) {
	h := newMux("abc123", true, newMetrics())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/hello", nil))
	if w.Code != 500 {
		t.Errorf("/hello with FAULT_HELLO: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/healthz", nil))
	if w.Code != 200 {
		t.Errorf("/healthz must stay healthy: %d", w.Code)
	}
}

func TestMetrics(t *testing.T) {
	m := newMetrics()
	h := m.instrument(newMux("v", true, m))
	for _, p := range []string{"/hello", "/hello", "/healthz", "/nope"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", p, nil))
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	out := w.Body.String()
	for _, want := range []string{
		`http_requests_total{route="GET /hello",code="500"} 2`,
		`http_requests_total{route="GET /healthz",code="200"} 1`,
		`http_requests_total{route="unmatched",code="404"} 1`,
		`http_request_duration_seconds_count{route="GET /hello"} 2`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s\n%s", want, out)
		}
	}
	if strings.Contains(out, "GET /metrics") {
		t.Error("/metrics itself must not be counted")
	}
}
