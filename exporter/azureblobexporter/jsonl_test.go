// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)

func TestMetricsJSONLMarshaler(t *testing.T) {
	marshaler := &metricsJSONLMarshaler{}
	md := testdata.GenerateMetricsTwoMetrics()

	result, err := marshaler.MarshalMetrics(md)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	assert.Greater(t, len(lines), 1)

	for _, line := range lines {
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "resourceMetrics")
	}
}

func TestTracesJSONLMarshaler(t *testing.T) {
	marshaler := &tracesJSONLMarshaler{}
	td := testdata.GenerateTracesTwoSpansSameResource()

	result, err := marshaler.MarshalTraces(td)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	assert.Greater(t, len(lines), 1)

	for _, line := range lines {
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "resourceSpans")
	}
}

func TestLogsJSONLMarshaler(t *testing.T) {
	marshaler := &logsJSONLMarshaler{}
	ld := testdata.GenerateLogsTwoLogRecordsSameResource()

	result, err := marshaler.MarshalLogs(ld)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	assert.Greater(t, len(lines), 1)

	for _, line := range lines {
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "resourceLogs")
	}
}

func TestTracesJSONLWithEmptyAttributes(t *testing.T) {
	td := testdata.GenerateTracesTwoSpansSameResource()
	rs := td.ResourceSpans().At(0)
	rs.Resource().Attributes().PutEmpty("empty_attr")
	rs.Resource().Attributes().PutStr("valid_attr", "value")

	ss := rs.ScopeSpans().At(0)
	ss.Scope().Attributes().PutEmpty("empty_scope")

	span := ss.Spans().At(0)
	span.Attributes().PutEmpty("empty_span")
	span.Events().AppendEmpty().Attributes().PutEmpty("empty_event")
	span.Links().AppendEmpty().Attributes().PutEmpty("empty_link")

	marshaler := &tracesJSONLMarshaler{}
	result, err := marshaler.MarshalTraces(td)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	for _, line := range lines {
		assert.NotContains(t, line, "null")
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		require.NoError(t, err)
	}
}

func TestCleanAttrs(t *testing.T) {
	attrs := pcommon.NewMap()
	attrs.PutStr("valid", "value")
	attrs.PutEmpty("empty")

	cleanAttrs(attrs)

	assert.Equal(t, 1, attrs.Len())
	value, exists := attrs.Get("valid")
	assert.True(t, exists)
	assert.Equal(t, "value", value.Str())

	_, exists = attrs.Get("empty")
	assert.False(t, exists)
}

func TestRemoveNullValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no nulls",
			input:    `{"key":"value"}`,
			expected: `{"key":"value"}`,
		},
		{
			name:     "with nulls",
			input:    `{"key":"value","null_key":null}`,
			expected: `{"key":"value"}`,
		},
		{
			name:     "trailing comma",
			input:    `{"null_key":null,"key":"value"}`,
			expected: `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeNullValues([]byte(tt.input))
			assert.JSONEq(t, tt.expected, string(result))
		})
	}
}

func TestTracesJSONLEmpty(t *testing.T) {
	td := ptrace.NewTraces()
	marshaler := &tracesJSONLMarshaler{}
	result, err := marshaler.MarshalTraces(td)
	require.NoError(t, err)
	assert.Empty(t, result)
}

