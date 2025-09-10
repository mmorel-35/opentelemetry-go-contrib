// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package skywalking

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	WithTracingMode        = withTracingMode
	TracingModeFromContext = tracingModeFromContext
)

func TestWithTracingMode(t *testing.T) {
	testCases := []struct {
		name string
		mode string
	}{
		{
			name: "normal mode",
			mode: TracingModeNormal,
		},
		{
			name: "skip analysis mode",
			mode: TracingModeSkipAnalysis,
		},
		{
			name: "custom mode",
			mode: "custom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := withTracingMode(context.Background(), tc.mode)
			mode := tracingModeFromContext(ctx)
			assert.Equal(t, tc.mode, mode)
		})
	}
}

func TestTracingModeFromContext_Default(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "nil context",
			ctx:  nil,
		},
		{
			name: "empty context",
			ctx:  context.Background(),
		},
		{
			name: "context with different value type",
			ctx:  context.WithValue(context.Background(), tracingModeKey, 123),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mode := tracingModeFromContext(tc.ctx)
			assert.Equal(t, TracingModeNormal, mode)
		})
	}
}

func TestTracingModeConstants(t *testing.T) {
	assert.Equal(t, "0", TracingModeNormal)
	assert.Equal(t, "1", TracingModeSkipAnalysis)
}
