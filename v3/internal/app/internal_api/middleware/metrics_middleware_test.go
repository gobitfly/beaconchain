package middleware

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeCounter struct {
	calls []struct{ method, code string }
}

func (f *fakeCounter) Inc(method, code string) {
	f.calls = append(f.calls, struct{ method, code string }{method, code})
}

func TestMetricsUnaryInterceptor_OK(t *testing.T) {
	f := &fakeCounter{}
	interceptor := MetricsUnaryInterceptor(f)

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{
		FullMethod: "/pkg.Svc/Do",
	}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.calls))
	}
	if f.calls[0].method != "/pkg.Svc/Do" { // leading slash is preserved by interceptor
		t.Errorf("method = %q, want %q", f.calls[0].method, "/pkg.Svc/Do")
	}
	if f.calls[0].code != "OK" {
		t.Errorf("code = %q, want %q", f.calls[0].code, "OK")
	}
}

func TestMetricsUnaryInterceptor_ErrorWithStatus(t *testing.T) {
	f := &fakeCounter{}
	interceptor := MetricsUnaryInterceptor(f)

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "nope")
	}

	_, _ = interceptor(context.Background(), "req", &grpc.UnaryServerInfo{
		FullMethod: "/pkg.Svc/Find",
	}, handler)

	if len(f.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.calls))
	}
	if f.calls[0].method != "/pkg.Svc/Find" {
		t.Errorf("method = %q, want %q", f.calls[0].method, "/pkg.Svc/Find")
	}
	if f.calls[0].code != "NotFound" { // matches codes.NotFound.String()
		t.Errorf("code = %q, want %q", f.calls[0].code, "NotFound")
	}
}

func TestMetricsUnaryInterceptor_ErrorUnknown(t *testing.T) {
	f := &fakeCounter{}
	interceptor := MetricsUnaryInterceptor(f)

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("boom")
	}

	_, _ = interceptor(context.Background(), "req", &grpc.UnaryServerInfo{
		FullMethod: "/pkg.Svc/Crash",
	}, handler)

	if len(f.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.calls))
	}
	if f.calls[0].method != "/pkg.Svc/Crash" {
		t.Errorf("method = %q, want %q", f.calls[0].method, "/pkg.Svc/Crash")
	}
	if f.calls[0].code != "Unknown" { // status.Code(err).String() for generic errors
		t.Errorf("code = %q, want %q", f.calls[0].code, "Unknown")
	}
}

func TestMetricsUnaryInterceptor_NilWriterPassthrough(t *testing.T) {
	// When writer is nil, interceptor should be a no-op wrapper around handler.
	interceptor := MetricsUnaryInterceptor(nil)

	called := false
	handler := func(ctx context.Context, req any) (any, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{
		FullMethod: "/pkg.Svc/Ping",
	}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("handler was not called")
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}
