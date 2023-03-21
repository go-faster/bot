package app

import (
	"context"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Resource returns new resource for application.
func Resource(ctx context.Context) (*resource.Resource, error) {
	opts := []resource.Option{
		resource.WithProcessRuntimeDescription(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeName(),
		resource.WithTelemetrySDK(),
	}
	r, err := resource.New(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "new")
	}
	return resource.Merge(resource.Default(), r)
}
