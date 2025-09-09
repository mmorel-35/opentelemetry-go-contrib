// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package skywalking_test

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/propagators/skywalking"
)

func ExamplePropagator() {
	// Create a new SkyWalking propagator
	skyWalkingPropagator := skywalking.Propagator{}

	// Set up the propagator in the global provider
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			skyWalkingPropagator,
			propagation.TraceContext{}, // Also support W3C trace context
			propagation.Baggage{},      // Also support baggage
		),
	)

	// Create a span context to propagate
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		log.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("0102030405060708")
	if err != nil {
		log.Fatal(err)
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	// Create a context with the span context
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

	// Inject the context into a carrier (e.g., HTTP headers)
	carrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	fmt.Printf("SkyWalking header set: %t\n", carrier.Get("sw8") != "")

	// Extract the context from the carrier
	extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
	extractedSC := trace.SpanContextFromContext(extractedCtx)

	fmt.Printf("Context extracted successfully: %t\n", extractedSC.IsValid())
	fmt.Printf("Trace ID preserved: %t\n", extractedSC.TraceID() == traceID)

	// Output:
	// SkyWalking header set: true
	// Context extracted successfully: true
	// Trace ID preserved: true
}

func ExampleNew() {
	// Create a configured SkyWalking propagator with service information
	skyWalkingPropagator := skywalking.New(
		skywalking.WithServiceName("my-service"),
		skywalking.WithServiceInstance("instance-1"),
		skywalking.WithEndpoint("/api/users"),
		skywalking.WithTargetAddress("downstream:8080"),
	)

	// Set up the propagator in the global provider
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			skyWalkingPropagator,
			propagation.TraceContext{}, // Also support W3C trace context
			propagation.Baggage{},      // Also support baggage
		),
	)

	// Create a span context to propagate
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		log.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("0102030405060708")
	if err != nil {
		log.Fatal(err)
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	// Create a context with the span context
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

	// Inject the context into a carrier (e.g., HTTP headers)
	carrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	fmt.Printf("SkyWalking header set: %t\n", carrier.Get("sw8") != "")

	// Extract the context from the carrier
	extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
	extractedSC := trace.SpanContextFromContext(extractedCtx)

	fmt.Printf("Context extracted successfully: %t\n", extractedSC.IsValid())
	fmt.Printf("Trace ID preserved: %t\n", extractedSC.TraceID() == traceID)

	// Output:
	// SkyWalking header set: true
	// Context extracted successfully: true
	// Trace ID preserved: true
}
