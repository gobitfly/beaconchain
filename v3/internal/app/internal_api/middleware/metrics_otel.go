package middleware

import (
	"context"
	"os"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type OtelCounter struct {
	projectID   string
	serviceName string
	revision    string

	record func(method, code string)
	stop   func(ctx context.Context) error
}

func NewOtelCounter(ctx context.Context, projectID, serviceName, revision string) (*OtelCounter, error) {
	exp, err := mexporter.New(
		mexporter.WithProjectID(projectID),
	)
	if err != nil {
		return nil, err
	}

	reader := sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(5*time.Second))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)

	meter := mp.Meter("api-internal")
	counter, err := meter.Int64Counter("request_count")
	if err != nil {
		_ = mp.Shutdown(ctx)
		return nil, err
	}

	svc := serviceName
	if svc == "" {
		svc = os.Getenv("K_SERVICE")
	}
	rev := revision
	if rev == "" {
		rev = os.Getenv("K_REVISION")
	}

	record := func(method, code string) {
		counter.Add(context.Background(), 1,
			otelmetric.WithAttributes(
				attribute.String("method", method),
				attribute.String("service_name", svc),
				attribute.String("code", code),
			),
		)
	}
	stop := func(ctx context.Context) error {
		return mp.Shutdown(ctx)
	}

	return &OtelCounter{
		projectID:   projectID,
		serviceName: svc,
		revision:    rev,
		record:      record,
		stop:        stop,
	}, nil
}

func (o *OtelCounter) Inc(method, code string) {
	o.record(method, code)
}

func (o *OtelCounter) Close(ctx context.Context) error {
	return o.stop(ctx)
}
