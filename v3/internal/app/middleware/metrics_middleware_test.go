package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeCounter struct {
	calls []struct{ method, code string }
}

func (f *fakeCounter) Inc(method, code string) {
	f.calls = append(f.calls, struct{ method, code string }{method, code})
}

func TestMetricsHTTPMiddleware_OK(t *testing.T) {
	f := &fakeCounter{}
	mw := MetricsHTTPMiddleware(f)

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/foo", nil)
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)

	if len(f.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.calls))
	}
	if f.calls[0].method != "GET" {
		t.Errorf("method = %q, want %q", f.calls[0].method, "GET")
	}
	if f.calls[0].code != "OK" {
		t.Errorf("code = %q, want %q", f.calls[0].code, "OK")
	}
}

func TestMetricsHTTPMiddleware_NotFound(t *testing.T) {
	f := &fakeCounter{}
	mw := MetricsHTTPMiddleware(f)

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest("POST", "/bar", nil)
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)

	if len(f.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.calls))
	}
	if f.calls[0].method != "POST" {
		t.Errorf("method = %q, want %q", f.calls[0].method, "POST")
	}
	if f.calls[0].code != "Not Found" {
		t.Errorf("code = %q, want %q", f.calls[0].code, "Not Found")
	}
}

func TestMetricsHTTPMiddleware_DefaultStatus(t *testing.T) {
	f := &fakeCounter{}
	mw := MetricsHTTPMiddleware(f)

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Don't call WriteHeader, should default to 200 OK
	})

	req := httptest.NewRequest("PUT", "/baz", nil)
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)

	if len(f.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.calls))
	}
	if f.calls[0].method != "PUT" {
		t.Errorf("method = %q, want %q", f.calls[0].method, "PUT")
	}
	if f.calls[0].code != "OK" {
		t.Errorf("code = %q, want %q", f.calls[0].code, "OK")
	}
}

func TestMetricsHTTPMiddleware_NilWriterPassthrough(t *testing.T) {
	mw := MetricsHTTPMiddleware(nil)

	called := false
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		called = true
		rw.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest("PATCH", "/qux", nil)
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)

	if !called {
		t.Fatalf("handler was not called")
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected response code: %v", rec.Code)
	}
}
