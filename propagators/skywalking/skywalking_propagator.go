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
	// SkyWalking v3 headers.
	sw8Header            = "sw8"
	sw8CorrelationHeader = "sw8-correlation"

	// Optional headers for service metadata (can be set by applications).
	sw8ServiceNameHeader     = "sw8-service-name"
	sw8ServiceInstanceHeader = "sw8-service-instance"
	sw8EndpointHeader        = "sw8-endpoint"
	sw8TargetAddressHeader   = "sw8-target-address"

	// Header field separator.
	fieldSeparator = "-"

	// SW8 header format (based on SkyWalking v3 specification):
	// sw8: {sample}-{trace-id}-{parent-trace-segment-id}-{parent-span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{target-address}
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

// Propagator implements the SkyWalking propagator.
//
// This propagator extracts and injects trace context using SkyWalking v3 headers.
// The sw8 header contains trace context information, while sw8-correlation can
// contain additional correlation data.
//
// For service metadata (service name, service instance, endpoint, target address),
// the propagator will first check for corresponding headers in the carrier
// (sw8-service-name, sw8-service-instance, sw8-endpoint, sw8-target-address).
// If not found, it will use default "unknown" values.
type Propagator struct{}

var _ propagation.TextMapPropagator = &Propagator{}

// Inject injects the trace context into the carrier using SkyWalking headers.
//
// This implementation follows the SkyWalking v3 specification for the sw8 header format:
// sw8: {sample}-{trace-id}-{parent-trace-segment-id}-{parent-span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{target-address}
//
// For service metadata fields (4-7), the propagator will first check for corresponding
// headers in the carrier. If not found, it will use default "unknown" values.
func (Propagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.TraceID().IsValid() || !sc.SpanID().IsValid() {
		return
	}

	// Determine sample flag according to spec: 0 or 1
	sampleFlag := sampleFlagNotSampled
	if sc.IsSampled() {
		sampleFlag = sampleFlagSampled
	}

	// Convert span ID to integer for field 3 (parent span ID)
	// Use the span ID's lower 64 bits as an integer, but ensure it's not negative
	var parentSpanID int64
	for i := 0; i < 8; i++ {
		parentSpanID = (parentSpanID << 8) | int64(sc.SpanID()[i])
	}
	// Ensure positive value
	if parentSpanID < 0 {
		parentSpanID = -parentSpanID
	}

	// Get service metadata from carrier, fall back to defaults if not present
	serviceName := carrier.Get(sw8ServiceNameHeader)
	if serviceName == "" {
		serviceName = unknownServiceName
	}
	serviceInstance := carrier.Get(sw8ServiceInstanceHeader)
	if serviceInstance == "" {
		serviceInstance = unknownServiceInstance
	}
	endpoint := carrier.Get(sw8EndpointHeader)
	if endpoint == "" {
		endpoint = unknownEndpoint
	}
	targetAddress := carrier.Get(sw8TargetAddressHeader)
	if targetAddress == "" {
		targetAddress = unknownAddress
	}

	// Build sw8 header according to SkyWalking v3 specification  
	// Format: {sample}-{trace-id}-{parent-trace-segment-id}-{parent-span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{target-address}
	sw8Value := strings.Join([]string{
		sampleFlag, // Field 0: sample (0 or 1)
		base64.StdEncoding.EncodeToString([]byte(sc.TraceID().String())),  // Field 1: trace ID (base64 encoded hex string)
		base64.StdEncoding.EncodeToString([]byte(sc.SpanID().String())),   // Field 2: parent trace segment ID (base64 encoded hex string)
		strconv.FormatInt(parentSpanID, 10),                               // Field 3: parent span ID (integer)
		base64.StdEncoding.EncodeToString([]byte(serviceName)),            // Field 4: parent service (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(serviceInstance)),        // Field 5: parent service instance (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(endpoint)),               // Field 6: parent endpoint (base64 encoded)
		base64.StdEncoding.EncodeToString([]byte(targetAddress)),          // Field 7: target address (base64 encoded)
	}, fieldSeparator)

	carrier.Set(sw8Header, sw8Value)
}

// Extract extracts the trace context from the carrier if it contains SkyWalking headers.
//
// This implementation follows the SkyWalking v3 specification for parsing the sw8 header.
func (Propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
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
// SW8 header format: {sample}-{trace-id}-{parent-trace-segment-id}-{parent-span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{target-address}
func extractFromSw8(sw8Value string) (trace.SpanContext, error) {
	fields := strings.Split(sw8Value, fieldSeparator)
	if len(fields) < expectedSw8Fields {
		return empty, errInsufficientFields
	}

	// Parse sample flag (field 0): 0 or 1
	sampleFlag := fields[0]
	isSampled := sampleFlag == sampleFlagSampled

	// Parse trace ID (field 1, base64 encoded hex string)
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

	// Parse parent trace segment ID (field 2, base64 encoded hex string) - use this as span ID for OpenTelemetry
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

	// Note: field 3 is the parent span ID as integer
	// Fields 4-7 contain service information (parent service, parent service instance, parent endpoint, target address)
	// These are not directly mapped to OpenTelemetry span context

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
func (Propagator) Fields() []string {
	return []string{sw8Header, sw8CorrelationHeader}
}
