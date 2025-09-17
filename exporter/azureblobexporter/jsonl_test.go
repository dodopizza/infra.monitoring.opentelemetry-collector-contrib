// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)

func TestMetricsJSONLMarshaler(t *testing.T) {
	marshaler := &metricsJSONLMarshaler{}
	md := testdata.GenerateMetricsTwoMetrics()

	result, err := marshaler.MarshalMetrics(md)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check that result contains multiple JSON lines
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	assert.Greater(t, len(lines), 1, "Expected multiple JSON lines")

	// Verify each line is valid JSON
	for i, line := range lines {
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		assert.NoErrorf(t, err, "Line %d is not valid JSON: %s", i, line)

		// Verify structure contains expected fields
		assert.Contains(t, parsed, "resourceMetrics", "Line %d missing resourceMetrics", i)
	}
}

func TestTracesJSONLMarshaler(t *testing.T) {
	marshaler := &tracesJSONLMarshaler{}
	td := testdata.GenerateTracesTwoSpansSameResource()

	result, err := marshaler.MarshalTraces(td)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check that result contains multiple JSON lines
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	assert.Greater(t, len(lines), 1, "Expected multiple JSON lines")

	// Verify each line is valid JSON
	for i, line := range lines {
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		assert.NoErrorf(t, err, "Line %d is not valid JSON: %s", i, line)

		// Verify structure contains expected fields
		assert.Contains(t, parsed, "resourceSpans", "Line %d missing resourceSpans", i)
	}
}

func TestLogsJSONLMarshalerCompatibility(t *testing.T) {
	// Test that existing logs JSONL marshaler still works
	marshaler := &logsJSONLMarshaler{}
	ld := testdata.GenerateLogsTwoLogRecordsSameResource()

	result, err := marshaler.MarshalLogs(ld)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check that result contains multiple JSON lines
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	assert.Greater(t, len(lines), 1, "Expected multiple JSON lines")

	// Verify each line is valid JSON
	for i, line := range lines {
		var parsed map[string]any
		err := json.Unmarshal([]byte(line), &parsed)
		assert.NoErrorf(t, err, "Line %d is not valid JSON: %s", i, line)

		// Verify structure contains expected fields
		assert.Contains(t, parsed, "resourceLogs", "Line %d missing resourceLogs", i)
	}
}

func TestJSONLFormatComparison(t *testing.T) {
	// Generate test data
	md := testdata.GenerateMetricsTwoMetrics()

	// Marshal with regular JSON
	jsonMarshaler := &pmetric.JSONMarshaler{}
	jsonResult, err := jsonMarshaler.MarshalMetrics(md)
	require.NoError(t, err)

	// Marshal with JSONL
	jsonlMarshaler := &metricsJSONLMarshaler{}
	jsonlResult, err := jsonlMarshaler.MarshalMetrics(md)
	require.NoError(t, err)

	// JSONL should be different from regular JSON (contains newlines)
	assert.NotEqual(t, jsonResult, jsonlResult)
	assert.Contains(t, string(jsonlResult), "\n", "JSONL should contain newlines")

	// Count lines in JSONL output
	scanner := bufio.NewScanner(bytes.NewReader(jsonlResult))
	lineCount := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			lineCount++
		}
	}
	assert.Positive(t, lineCount, "Should have at least one line")
}

func TestRemoveEmptyAttributesFromTraces(t *testing.T) {
	// Test that empty attributes are properly removed to prevent Jaeger errors
	td := testdata.GenerateTracesTwoSpansSameResource()

	// Add some empty attributes to simulate real-world scenario
	rs := td.ResourceSpans().At(0)
	rs.Resource().Attributes().PutEmpty("empty_resource_attr")
	rs.Resource().Attributes().PutStr("valid_resource_attr", "value")

	ss := rs.ScopeSpans().At(0)
	ss.Scope().Attributes().PutEmpty("empty_scope_attr")
	ss.Scope().Attributes().PutStr("valid_scope_attr", "value")

	span := ss.Spans().At(0)
	span.Attributes().PutEmpty("empty_span_attr")
	span.Attributes().PutStr("valid_span_attr", "value")

	// Count empty attributes before cleaning
	emptyCountBefore := 0
	rs.Resource().Attributes().Range(func(_ string, v pcommon.Value) bool {
		if v.Type() == pcommon.ValueTypeEmpty {
			emptyCountBefore++
		}
		return true
	})
	ss.Scope().Attributes().Range(func(_ string, v pcommon.Value) bool {
		if v.Type() == pcommon.ValueTypeEmpty {
			emptyCountBefore++
		}
		return true
	})
	span.Attributes().Range(func(_ string, v pcommon.Value) bool {
		if v.Type() == pcommon.ValueTypeEmpty {
			emptyCountBefore++
		}
		return true
	})

	assert.Equal(t, 3, emptyCountBefore, "Should have 3 empty attributes before cleaning")

	// Clean empty attributes
	removeEmptyAttributesFromTraces(td)

	// Verify no empty attributes remain
	emptyCountAfter := 0
	rs.Resource().Attributes().Range(func(_ string, v pcommon.Value) bool {
		if v.Type() == pcommon.ValueTypeEmpty {
			emptyCountAfter++
		}
		return true
	})
	ss.Scope().Attributes().Range(func(_ string, v pcommon.Value) bool {
		if v.Type() == pcommon.ValueTypeEmpty {
			emptyCountAfter++
		}
		return true
	})
	span.Attributes().Range(func(_ string, v pcommon.Value) bool {
		if v.Type() == pcommon.ValueTypeEmpty {
			emptyCountAfter++
		}
		return true
	})

	assert.Equal(t, 0, emptyCountAfter, "Should have 0 empty attributes after cleaning")

	// Verify valid attributes are preserved
	validValue, exists := rs.Resource().Attributes().Get("valid_resource_attr")
	assert.True(t, exists)
	assert.Equal(t, "value", validValue.Str())
}

