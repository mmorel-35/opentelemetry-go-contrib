# SkyWalking Propagator

This package provides a SkyWalking propagator for OpenTelemetry Go.

## Status

⚠️ **INCOMPLETE IMPLEMENTATION** ⚠️

This is a foundational implementation of the SkyWalking propagator that follows the established patterns used by other propagators in this repository. However, it requires completion based on the official SkyWalking specification.

## What's Implemented

- [x] Basic project structure following OpenTelemetry Go Contrib patterns
- [x] Go module setup with proper dependencies
- [x] Implementation of `propagation.TextMapPropagator` interface
- [x] Basic header injection and extraction for `sw8` and `sw8-correlation` headers
- [x] Comprehensive test suite with unit tests and benchmarks
- [x] Example usage documentation
- [x] Proper error handling structure
- [x] Version management

## What's Missing (Required to Complete)

The implementation is currently incomplete and requires the following information from the official SkyWalking specification:

### 1. SW8 Header Format Specification
- **Current**: Placeholder format with basic fields
- **Needed**: Exact field order, encoding, and format requirements
- **Reference**: https://skywalking.apache.org/docs/main/next/en/api/x-process-propagation-headers-v3/

### 2. Field Encoding Details
- **Trace ID encoding**: Format and padding requirements
- **Span ID encoding**: Format and relationship to segment ID
- **Service information**: How to encode service name, instance, and endpoint
- **Sampling flags**: How sampling decisions are represented

### 3. SW8-Correlation Header (if supported)
- **Format**: Key-value encoding for correlation data
- **Limits**: Maximum size, allowed characters, etc.
- **Integration**: How it relates to OpenTelemetry baggage

### 4. Error Handling Requirements
- **Invalid headers**: How to handle malformed headers
- **Backward compatibility**: Support for older header versions
- **Fallback behavior**: What to do when headers are partially invalid

### 5. Integration Details
- **Service mapping**: How OpenTelemetry service concepts map to SkyWalking
- **Resource attributes**: Which resource attributes should be used for service info
- **Sampling**: How OpenTelemetry sampling decisions map to SkyWalking flags

## Current Placeholder Implementation

The current implementation uses a basic 9-field format for the sw8 header:
```
sw8: {trace-id}-{segment-id}-{span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{target-service}-{target-service-instance}-{target-endpoint}
```

This is based on general knowledge of SkyWalking but needs verification and proper implementation according to the official specification.

## Usage

```go
import "go.opentelemetry.io/contrib/propagators/skywalking"

// Create propagator
propagator := skywalking.SkyWalking{}

// Use with OpenTelemetry
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagator,
    propagation.TraceContext{},
    propagation.Baggage{},
))
```

## Testing

```bash
# Run tests
go test ./...

# Run benchmarks
go test -bench=.

# Check test coverage
go test -cover ./...
```

## Next Steps

1. **Obtain SkyWalking specification**: Access the official documentation to get exact header format requirements
2. **Implement proper encoding**: Update the inject/extract methods with correct field encoding
3. **Add service mapping**: Implement proper service information extraction from OpenTelemetry context
4. **Update tests**: Add comprehensive tests for various header formats and edge cases
5. **Add integration tests**: Test interoperability with actual SkyWalking agents
6. **Documentation**: Add comprehensive usage examples and migration guides

## Dependencies

- Go 1.24+
- OpenTelemetry Go v1.38.0+

## License

Apache 2.0 - See LICENSE file for details.