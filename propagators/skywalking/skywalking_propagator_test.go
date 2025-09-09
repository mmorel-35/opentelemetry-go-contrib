// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package skywalking

import (
	"context"
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

func TestSkyWalkingPropagator_Interface(t *testing.T) {
	var _ propagation.TextMapPropagator = &SkyWalking{}
}

func TestSkyWalkingPropagator_Fields(t *testing.T) {
	p := SkyWalking{}
	fields := p.Fields()
	
	assert.Contains(t, fields, sw8Header)
	assert.Contains(t, fields, sw8CorrelationHeader)
	assert.Len(t, fields, 2)
}

func TestSkyWalkingPropagator_Inject_EmptyContext(t *testing.T) {
	p := SkyWalking{}
	carrier := make(propagation.MapCarrier)
	
	// Inject with empty context should not set any headers
	p.Inject(context.Background(), carrier)
	
	assert.Empty(t, carrier.Get(sw8Header))
}

func TestSkyWalkingPropagator_Inject_ValidContext(t *testing.T) {
	p := SkyWalking{}
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
	
	// The header should contain the trace and span IDs
	assert.Contains(t, sw8Value, traceID.String())
	assert.Contains(t, sw8Value, spanID.String())
}

func TestSkyWalkingPropagator_Extract_EmptyCarrier(t *testing.T) {
	p := SkyWalking{}
	carrier := make(propagation.MapCarrier)
	
	ctx := p.Extract(context.Background(), carrier)
	
	// Should return the original context
	assert.Equal(t, context.Background(), ctx)
}

func TestSkyWalkingPropagator_Extract_InvalidHeader(t *testing.T) {
	p := SkyWalking{}
	carrier := make(propagation.MapCarrier)
	
	// Set an invalid sw8 header
	carrier.Set(sw8Header, "invalid-header")
	
	ctx := p.Extract(context.Background(), carrier)
	
	// Should return the original context
	assert.Equal(t, context.Background(), ctx)
}

func TestSkyWalkingPropagator_RoundTrip(t *testing.T) {
	p := SkyWalking{}
	
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

// TestSkyWalkingPropagator_ExtractWithMinimalHeader tests extraction with a minimal valid header
func TestSkyWalkingPropagator_ExtractWithMinimalHeader(t *testing.T) {
	p := SkyWalking{}
	carrier := make(propagation.MapCarrier)
	
	// Create a minimal valid sw8 header (placeholder format)
	sw8Value := traceID.String() + "-" + spanID.String() + "-" + spanID.String() + 
		"-service-instance-endpoint-target-target-instance"
	carrier.Set(sw8Header, sw8Value)
	
	ctx := p.Extract(context.Background(), carrier)
	sc := trace.SpanContextFromContext(ctx)
	
	require.True(t, sc.IsValid())
	assert.Equal(t, traceID, sc.TraceID())
	assert.Equal(t, spanID, sc.SpanID())
}

// Benchmark tests
func BenchmarkSkyWalkingPropagator_Inject(b *testing.B) {
	p := SkyWalking{}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		carrier := make(propagation.MapCarrier)
		p.Inject(ctx, carrier)
	}
}

func BenchmarkSkyWalkingPropagator_Extract(b *testing.B) {
	p := SkyWalking{}
	carrier := make(propagation.MapCarrier)
	sw8Value := traceID.String() + "-" + spanID.String() + "-" + spanID.String() + 
		"-service-instance-endpoint-target-target-instance"
	carrier.Set(sw8Header, sw8Value)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Extract(context.Background(), carrier)
	}
}