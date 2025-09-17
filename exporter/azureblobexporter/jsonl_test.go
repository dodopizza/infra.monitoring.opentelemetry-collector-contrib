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

func TestCleanAttrs(t *testing.T) {
	// Test that cleanAttrs properly removes empty attributes
	attrs := pcommon.NewMap()
	attrs.PutStr("valid_string", "value")
	attrs.PutInt("valid_int", 123)
	attrs.PutEmpty("empty_value")

	// Verify we have 3 attributes before cleaning
	assert.Equal(t, 3, attrs.Len(), "Should have 3 attributes before cleaning")

	// Clean empty attributes
	cleanAttrs(attrs)

	// Verify only 2 attributes remain (empty one removed)
	assert.Equal(t, 2, attrs.Len(), "Should have 2 attributes after cleaning")

	// Verify no empty attributes remain
	attrs.Range(func(k string, v pcommon.Value) bool {
		assert.NotEqual(t, pcommon.ValueTypeEmpty, v.Type(), "No empty values should remain: %s", k)
		return true
	})

	// Verify valid attributes are preserved
	validValue, exists := attrs.Get("valid_string")
	assert.True(t, exists)
	assert.Equal(t, "value", validValue.Str())

	validInt, exists := attrs.Get("valid_int")
	assert.True(t, exists)
	assert.Equal(t, int64(123), validInt.Int())
}

