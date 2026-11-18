package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMux(t *testing.T) {
	h := newMux("abc123", false)
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
	h := newMux("abc123", true)
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
