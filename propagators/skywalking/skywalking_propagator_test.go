// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package skywalking

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	traceID = trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID  = trace.SpanID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
)

func TestSkyWalkingPropagator_Interface(_ *testing.T) {
	var _ propagation.TextMapPropagator = &Propagator{}
}

func TestSkyWalkingPropagator_Fields(t *testing.T) {
	p := Propagator{}
	fields := p.Fields()

	assert.Contains(t, fields, sw8Header)
	assert.Contains(t, fields, sw8CorrelationHeader)
	assert.Len(t, fields, 2)
}

func TestSkyWalkingPropagator_Inject_EmptyContext(t *testing.T) {
	p := Propagator{}
	carrier := make(propagation.MapCarrier)

	// Inject with empty context should not set any headers
	p.Inject(context.Background(), carrier)

	assert.Empty(t, carrier.Get(sw8Header))
}

func TestSkyWalkingPropagator_Inject_ValidContext(t *testing.T) {
	p := Propagator{}
	carrier := make(propagation.MapCarrier)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

	p.Inject(ctx, carrier)

	// Should set the sw8 header
	sw8Value := carrier.Get(sw8Header)
	assert.NotEmpty(t, sw8Value)

	// The header should be in the correct format with base64 encoded fields
	// Check that it starts with "1" (sampled flag) and has the right number of fields
	fields := strings.Split(sw8Value, "-")
	assert.Len(t, fields, 8, "sw8 header should have 8 fields")
	assert.Equal(t, "1", fields[0], "first field should be sample flag = 1")
}

func TestSkyWalkingPropagator_Extract_EmptyCarrier(t *testing.T) {
	p := Propagator{}
	carrier := make(propagation.MapCarrier)

	ctx := p.Extract(context.Background(), carrier)

	// Should return the original context
	assert.Equal(t, context.Background(), ctx)
}

func TestSkyWalkingPropagator_Extract_InvalidHeader(t *testing.T) {
	p := Propagator{}
	carrier := make(propagation.MapCarrier)

	// Set an invalid sw8 header
	carrier.Set(sw8Header, "invalid-header")

	ctx := p.Extract(context.Background(), carrier)

	// Should return the original context
	assert.Equal(t, context.Background(), ctx)
}

func TestSkyWalkingPropagator_RoundTrip(t *testing.T) {
	p := Propagator{}

	// Create a span context
	originalSC := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	originalCtx := trace.ContextWithRemoteSpanContext(context.Background(), originalSC)

	// Inject into carrier
	carrier := make(propagation.MapCarrier)
	p.Inject(originalCtx, carrier)

	// Extract from carrier
	extractedCtx := p.Extract(context.Background(), carrier)
	extractedSC := trace.SpanContextFromContext(extractedCtx)

	// TODO: This test will need adjustment once the exact specification is implemented
	// For now, we just verify that some context was extracted
	assert.True(t, extractedSC.IsValid(), "extracted span context should be valid")

	// The trace ID should be preserved
	assert.Equal(t, originalSC.TraceID(), extractedSC.TraceID())

	// TODO: Verify other fields once the specification is complete
}

// TestSkyWalkingPropagator_ExtractWithMinimalHeader tests extraction with a minimal valid header.
func TestSkyWalkingPropagator_ExtractWithMinimalHeader(t *testing.T) {
	p := Propagator{}
	carrier := make(propagation.MapCarrier)

	// Create a minimal valid sw8 header in the correct format
	// Format: {sample-flag}-{trace-id}-{segment-id}-{span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{address-used-at-client}
	sw8Value := strings.Join([]string{
		"1", // sample flag
		base64.StdEncoding.EncodeToString([]byte(traceID.String())), // trace ID
		base64.StdEncoding.EncodeToString([]byte(spanID.String())),  // segment ID
		"123", // span ID as integer
		base64.StdEncoding.EncodeToString([]byte("test-service")),  // parent service
		base64.StdEncoding.EncodeToString([]byte("test-instance")), // parent service instance
		base64.StdEncoding.EncodeToString([]byte("test-endpoint")), // parent endpoint
		base64.StdEncoding.EncodeToString([]byte("test-address")),  // address
	}, "-")
	carrier.Set(sw8Header, sw8Value)

	ctx := p.Extract(context.Background(), carrier)
	sc := trace.SpanContextFromContext(ctx)

	require.True(t, sc.IsValid())
	assert.Equal(t, traceID, sc.TraceID())
	assert.Equal(t, spanID, sc.SpanID())
	assert.True(t, sc.IsSampled(), "should be sampled based on sample flag")
}

// Benchmark tests.
func BenchmarkSkyWalkingPropagator_Inject(b *testing.B) {
	p := Propagator{}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

	b.ResetTimer()
	for range b.N {
		carrier := make(propagation.MapCarrier)
		p.Inject(ctx, carrier)
	}
}

func BenchmarkSkyWalkingPropagator_Extract(b *testing.B) {
	p := Propagator{}
	carrier := make(propagation.MapCarrier)
	sw8Value := strings.Join([]string{
		"1", // sample flag
		base64.StdEncoding.EncodeToString([]byte(traceID.String())), // trace ID
		base64.StdEncoding.EncodeToString([]byte(spanID.String())),  // segment ID
		"123", // span ID as integer
		base64.StdEncoding.EncodeToString([]byte("service")),  // parent service
		base64.StdEncoding.EncodeToString([]byte("instance")), // parent service instance
		base64.StdEncoding.EncodeToString([]byte("endpoint")), // parent endpoint
		base64.StdEncoding.EncodeToString([]byte("target")),   // address
	}, "-")
	carrier.Set(sw8Header, sw8Value)

	b.ResetTimer()
	for range b.N {
		p.Extract(context.Background(), carrier)
	}
}

// Test configuration options.
func TestSkyWalkingPropagator_New(t *testing.T) {
	// Test New function with options
	p := New(
		WithServiceName("test-service"),
		WithServiceInstance("test-instance"),
		WithEndpoint("test-endpoint"),
		WithTargetAddress("127.0.0.1:8080"),
	)

	assert.Equal(t, "test-service", p.serviceName)
	assert.Equal(t, "test-instance", p.serviceInstance)
	assert.Equal(t, "test-endpoint", p.endpoint)
	assert.Equal(t, "127.0.0.1:8080", p.targetAddress)
}

func TestSkyWalkingPropagator_New_DefaultValues(t *testing.T) {
	// Test New function with no options uses defaults
	p := New()

	assert.Equal(t, unknownServiceName, p.serviceName)
	assert.Equal(t, unknownServiceInstance, p.serviceInstance)
	assert.Equal(t, unknownEndpoint, p.endpoint)
	assert.Equal(t, unknownAddress, p.targetAddress)
}

func TestSkyWalkingPropagator_ConfiguredInject(t *testing.T) {
	p := New(
		WithServiceName("my-service"),
		WithServiceInstance("my-instance"),
		WithEndpoint("/api/test"),
		WithTargetAddress("downstream:9090"),
	)
	carrier := make(propagation.MapCarrier)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

	p.Inject(ctx, carrier)

	sw8Value := carrier.Get(sw8Header)
	assert.NotEmpty(t, sw8Value, "sw8 header should be set")

	// Parse the sw8 header to verify configured values are used
	fields := strings.Split(sw8Value, "-")
	require.Len(t, fields, 8, "sw8 header should have 8 fields")

	// Check that configured values are properly base64 encoded in the header
	serviceBytes, err := base64.StdEncoding.DecodeString(fields[4])
	require.NoError(t, err)
	assert.Equal(t, "my-service", string(serviceBytes))

	instanceBytes, err := base64.StdEncoding.DecodeString(fields[5])
	require.NoError(t, err)
	assert.Equal(t, "my-instance", string(instanceBytes))

	endpointBytes, err := base64.StdEncoding.DecodeString(fields[6])
	require.NoError(t, err)
	assert.Equal(t, "/api/test", string(endpointBytes))

	addressBytes, err := base64.StdEncoding.DecodeString(fields[7])
	require.NoError(t, err)
	assert.Equal(t, "downstream:9090", string(addressBytes))
}
