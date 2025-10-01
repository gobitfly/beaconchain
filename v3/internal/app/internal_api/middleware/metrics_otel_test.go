//go:build testhooks

package middleware

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestOtelCounter_Inc_recordsPoint(t *testing.T) {
	// In-memory reader collects data without any exporter/network.
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	oc, err := newOtelCounterWithProvider(context.Background(), mp, "svc", "rev")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Call the wrapper.
	oc.Inc("pkg.Svc/Do", "OK")

	// Collect in-memory data.
	var out metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &out); err != nil {
		t.Fatalf("collect: %v", err)
	}

	// Find our counter datapoint and assert it incremented by 1.
	found := false
	for _, sm := range out.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "request_count" {
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok {
					t.Fatalf("metric not int64 sum")
				}
				if len(sum.DataPoints) == 0 {
					t.Fatalf("no datapoints")
				}
				// We only added one, so value should be 1.
				if sum.DataPoints[0].Value != 1 {
					t.Fatalf("got %d, want 1", sum.DataPoints[0].Value)
				}
				// Attributes are not attached in test seam (we keep prod code unchanged),
				// but this verifies the wrapper’s Add path is wired.
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("request_count not found")
	}

	_ = oc.Close(context.Background())
}

// Optional: ensure meter exists and counter can be created with the real provider surface.
func TestOtelCounter_API_shape(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	m := mp.Meter("api-internal")
	if _, err := m.Int64Counter("request_count", metric.WithDescription("test")); err != nil {
		t.Fatalf("counter: %v", err)
	}
}
