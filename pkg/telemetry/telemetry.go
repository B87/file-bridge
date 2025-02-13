package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SimpleSpanExporter is a custom span exporter that prints spans in a simple format
type SimpleSpanExporter struct{}

func (e *SimpleSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	for _, span := range spans {
		fmt.Printf("\n=== SPAN: %s ===\n", span.Name())
		fmt.Printf("TraceID: %s\n", span.SpanContext().TraceID())
		fmt.Printf("SpanID: %s\n", span.SpanContext().SpanID())

		// Print attributes with proper value formatting
		fmt.Println("Attributes:")
		for _, attr := range span.Attributes() {
			switch attr.Value.Type() {
			case attribute.STRING:
				fmt.Printf("  %s: %s\n", attr.Key, attr.Value.AsString())
			case attribute.BOOL:
				fmt.Printf("  %s: %t\n", attr.Key, attr.Value.AsBool())
			case attribute.INT64:
				fmt.Printf("  %s: %d\n", attr.Key, attr.Value.AsInt64())
			case attribute.FLOAT64:
				fmt.Printf("  %s: %f\n", attr.Key, attr.Value.AsFloat64())
			default:
				fmt.Printf("  %s: %v\n", attr.Key, attr.Value)
			}
		}

		// Print events with proper formatting
		if len(span.Events()) > 0 {
			fmt.Println("Events:")
			for _, event := range span.Events() {
				fmt.Printf("  %s:\n", event.Name)
				for _, attr := range event.Attributes {
					fmt.Printf("    %s: %v\n", attr.Key, attr.Value)
				}
			}
		}
		fmt.Println("================")
	}
	return nil
}

func (e *SimpleSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}
