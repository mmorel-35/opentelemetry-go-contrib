// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package skywalking // import "go.opentelemetry.io/contrib/propagators/skywalking"

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// SkyWalking v8 headers.
	sw8Header            = "sw8"
	sw8CorrelationHeader = "sw8-correlation"

	// Header field separator.
	fieldSeparator = "-"

	// SW8 header format (based on SkyWalking v8 specification):
	// sw8: {sample-flag}-{trace-id}-{segment-id}-{span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{address-used-at-client}
	expectedSw8Fields = 8

	// Sample flags.
	sampleFlagSampled    = "1"
	sampleFlagNotSampled = "0"

	// Default values for unknown fields.
	unknownServiceName     = "unknown"
	unknownServiceInstance = "unknown"
	unknownEndpoint        = "unknown"
	unknownAddress         = "unknown"
)

var (
	empty = trace.SpanContext{}

	// Error definitions.
	errInvalidTraceID     = errors.New("invalid trace ID in sw8 header")
	errInvalidSpanID      = errors.New("invalid span ID in sw8 header")
	errInsufficientFields = errors.New("insufficient fields in sw8 header")
	errBase64Decode       = errors.New("failed to decode base64 field")
)

// SkyWalking implements the SkyWalking propagator.
//
// This propagator extracts and injects trace context using SkyWalking v8 headers.
// The sw8 header contains trace context information, while sw8-correlation can
// contain additional correlation data.
type SkyWalking struct{}

var _ propagation.TextMapPropagator = &SkyWalking{}

// Inject injects the trace context into the carrier using SkyWalking headers.
//
// This implementation follows the SkyWalking v8 specification for the sw8 header format:
// sw8: {sample-flag}-{trace-id}-{segment-id}-{span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{address-used-at-client}
func (SkyWalking) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.TraceID().IsValid() || !sc.SpanID().IsValid() {
		return
	}

	// Determine sample flag
	sampleFlag := sampleFlagNotSampled
	if sc.IsSampled() {
		sampleFlag = sampleFlagSampled
	}

	// Build sw8 header according to specification
	// Format: {sample-flag}-{trace-id}-{segment-id}-{span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{address-used-at-client}
	sw8Value := strings.Join([]string{
		sampleFlag, // Field 0: sample flag
		base64.StdEncoding.EncodeToString([]byte(sc.TraceID().String())),  // Field 1: trace ID (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(sc.SpanID().String())),   // Field 2: segment ID (using span ID, base64 encoded)
		strconv.Itoa(int(sc.SpanID()[7])),                                 // Field 3: span ID (as integer)
		base64.StdEncoding.EncodeToString([]byte(unknownServiceName)),     // Field 4: parent service (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(unknownServiceInstance)), // Field 5: parent service instance (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(unknownEndpoint)),        // Field 6: parent endpoint (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(unknownAddress)),         // Field 7: address used at client (base64 encoded)
	}, fieldSeparator)

	carrier.Set(sw8Header, sw8Value)

	// TODO: Handle sw8-correlation header for baggage/correlation context if needed
}

// Extract extracts the trace context from the carrier if it contains SkyWalking headers.
//
// This implementation follows the SkyWalking v8 specification for parsing the sw8 header.
func (SkyWalking) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sw8Value := carrier.Get(sw8Header)
	if sw8Value == "" {
		return ctx
	}

	sc, err := extractFromSw8(sw8Value)
	if err != nil || !sc.IsValid() {
		return ctx
	}

	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

// extractFromSw8 extracts trace context from sw8 header value.
//
// SW8 header format: {sample-flag}-{trace-id}-{segment-id}-{span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{address-used-at-client}.
func extractFromSw8(sw8Value string) (trace.SpanContext, error) {
	fields := strings.Split(sw8Value, fieldSeparator)
	if len(fields) < expectedSw8Fields {
		return empty, errInsufficientFields
	}

	// Parse sample flag (field 0)
	sampleFlag := fields[0]
	isSampled := sampleFlag == sampleFlagSampled

	// Parse trace ID (field 1, base64 encoded)
	traceIDBytes, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil {
		return empty, errBase64Decode
	}
	traceIDStr := string(traceIDBytes)
	if traceIDStr == "" {
		return empty, errInvalidTraceID
	}

	traceID, err := trace.TraceIDFromHex(traceIDStr)
	if err != nil {
		return empty, errInvalidTraceID
	}

	// Parse segment ID (field 2, base64 encoded) - we'll use this to derive the span ID
	segmentIDBytes, err := base64.StdEncoding.DecodeString(fields[2])
	if err != nil {
		return empty, errBase64Decode
	}
	segmentIDStr := string(segmentIDBytes)
	if segmentIDStr == "" {
		return empty, errInvalidSpanID
	}

	spanID, err := trace.SpanIDFromHex(segmentIDStr)
	if err != nil {
		return empty, errInvalidSpanID
	}

	// Note: field 3 is the span ID as integer, but we use the segment ID (field 2) for OpenTelemetry span ID
	// Fields 4-7 contain service information that we don't currently map to OpenTelemetry context

	// Build span context
	var flags trace.TraceFlags
	if isSampled {
		flags = trace.FlagsSampled
	}

	scc := trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: flags,
	}

	return trace.NewSpanContext(scc), nil
}

// Fields returns the keys whose values are set with Inject.
func (SkyWalking) Fields() []string {
	return []string{sw8Header, sw8CorrelationHeader}
}
