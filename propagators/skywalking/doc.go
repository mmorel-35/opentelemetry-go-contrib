// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package skywalking implements the SkyWalking propagator specification.
// 
// SkyWalking uses sw8 headers for cross-process propagation of trace context.
// The propagator extracts and injects trace context using the SkyWalking v8 format.
//
// For more information about SkyWalking propagation, see:
// https://skywalking.apache.org/docs/main/next/en/api/x-process-propagation-headers-v3/
package skywalking // import "go.opentelemetry.io/contrib/propagators/skywalking"