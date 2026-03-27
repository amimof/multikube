package app

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.GetTracerProvider().Tracer("multikube-server")
