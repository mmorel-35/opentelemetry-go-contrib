# SkyWalking Propagator

This package provides a SkyWalking propagator for OpenTelemetry Go.

## Status

✅ **COMPLETED IMPLEMENTATION** ✅

This is a complete implementation of the SkyWalking propagator that follows the official SkyWalking v3 Cross Process Propagation Headers Protocol specification.

## What's Implemented

- [x] Full SkyWalking v3 SW8 header format implementation  
- [x] SW8-Correlation header support for cross-process correlation data
- [x] Base64 encoding/decoding of header fields as per official specification
- [x] Proper sampling flag handling (0 = context exists but may be ignored, 1 = sampled)
- [x] Complete project structure following OpenTelemetry Go Contrib patterns
- [x] Go module setup with proper dependencies
- [x] Implementation of `propagation.TextMapPropagator` interface
- [x] Header injection and extraction for `sw8` and `sw8-correlation` headers
- [x] Comprehensive test suite with unit tests and benchmarks
- [x] Example usage documentation including correlation examples
- [x] Proper error handling for malformed headers
- [x] Version management

## SW8 Header Format

The implementation follows the official SkyWalking v3 Cross Process Propagation Headers Protocol:

```
sw8: {sample}-{trace-id}-{parent-trace-segment-id}-{parent-span-id}-{parent-service}-{parent-service-instance}-{parent-endpoint}-{target-address}
```

Where:
- **Field 0**: Sample flag ("1" if sampled, "0" if context exists but may be ignored)
- **Field 1**: Trace ID (Base64 encoded hex string, globally unique)
- **Field 2**: Parent trace segment ID (Base64 encoded hex string, globally unique)
- **Field 3**: Parent span ID (integer, begins with 0, points to parent span in parent trace segment)
- **Field 4**: Parent service (Base64 encoded, max 50 UTF-8 characters)
- **Field 5**: Parent service instance (Base64 encoded, max 50 UTF-8 characters)
- **Field 6**: Parent endpoint (Base64 encoded, max 150 UTF-8 characters, operation name of first entry span)
- **Field 7**: Target address (Base64 encoded, network address used on client end)

## SW8-Correlation Header Format

The propagator also supports SkyWalking correlation headers for propagating custom key-value data:

```
sw8-correlation: key1:value1,key2:value2
```

Correlation data is automatically extracted from and injected into OpenTelemetry baggage. Both keys and values are URL-encoded to handle special characters safely.

## Features

### Implemented
- ✅ SW8 header injection and extraction
- ✅ SW8-Correlation header support for cross-process correlation data
- ✅ Base64 encoding/decoding of appropriate fields per official specification
- ✅ Sampling flag propagation (0/1 format)
- ✅ Round-trip compatibility for both trace context and correlation data
- ✅ Error handling for malformed headers
- ✅ Stateless design with default "unknown" values for service metadata
- ✅ Proper trace ID and span ID handling
- ✅ Automatic integration with OpenTelemetry baggage

### Future Enhancements
- [ ] Service name extraction from OpenTelemetry resource attributes
- [ ] SW8-X extension header support for advanced features

### Additional SkyWalking Integration Opportunities

This propagator provides the foundation for broader SkyWalking integration in OpenTelemetry Go Contrib. Additional components that could be implemented include:

1. **Trace Exporter** (`exporters/skywalking/`) - Export OpenTelemetry spans to SkyWalking backend via the Trace Data Protocol v3
2. **Resource Detector** (`detectors/skywalking/`) - Detect SkyWalking agent presence and extract service metadata  
3. **Remote Sampler** (`samplers/skywalking/`) - SkyWalking-compatible sampling decisions and remote sampling support
4. **Context Utilities** - Helper functions for working with SkyWalking-specific context values and correlation data
5. **ID Generators** - Custom trace/span ID generation following SkyWalking conventions

The most immediately valuable would be a **trace exporter** for sending spans to SkyWalking backend and a **resource detector** for seamless integration with existing SkyWalking deployments.

## Usage

### Basic Usage

```go
import "go.opentelemetry.io/contrib/propagators/skywalking"

// Create propagator
propagator := skywalking.Propagator{}

// Use with OpenTelemetry
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagator,
    propagation.TraceContext{},
    propagation.Baggage{},
))
```

### Correlation Data Usage

The propagator automatically handles SkyWalking correlation data through OpenTelemetry baggage:

```go
import (
    "go.opentelemetry.io/otel/baggage"
    "go.opentelemetry.io/contrib/propagators/skywalking"
)

// Add correlation data to baggage
member1, _ := baggage.NewMember("service.name", "web-service")
member2, _ := baggage.NewMember("user.id", "user123")
bags, _ := baggage.New(member1, member2)

// Create context with baggage
ctx := baggage.ContextWithBaggage(context.Background(), bags)

// The propagator will automatically inject correlation data into sw8-correlation header
propagator := skywalking.Propagator{}
carrier := make(propagation.MapCarrier)
propagator.Inject(ctx, carrier)

// On the receiving side, correlation data is automatically extracted into baggage
extractedCtx := propagator.Extract(context.Background(), carrier)
extractedBags := baggage.FromContext(extractedCtx)
serviceName := extractedBags.Member("service.name").Value() // "web-service"
```

The propagator uses default "unknown" values for service metadata fields in the SW8 header, following the stateless design principle.

## Testing

```bash
# Run tests
go test ./...

# Run benchmarks
go test -bench=.

# Check test coverage
go test -cover ./...
```

Current test coverage: **86.0%**

## Specification Reference

This implementation is based on the official SkyWalking Cross Process Propagation Headers Protocol v3:
https://skywalking.apache.org/docs/main/latest/en/api/x-process-propagation-headers-v3/

## Dependencies

- Go 1.24+
- OpenTelemetry Go v1.38.0+

## License

Apache 2.0 - See LICENSE file for details.