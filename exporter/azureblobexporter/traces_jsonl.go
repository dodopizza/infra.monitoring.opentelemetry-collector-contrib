// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bytes"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

type tracesJSONLMarshaler struct{}

func (*tracesJSONLMarshaler) MarshalTraces(td ptrace.Traces) ([]byte, error) {
	marshaler := &ptrace.JSONMarshaler{}
	var buf bytes.Buffer

	// Clean empty attributes once to prevent Jaeger "invalid tag type in <nil>" errors
	// This is much faster than per-span reconstruction for millions of spans
	cleanedTraces := ptrace.NewTraces()
	td.CopyTo(cleanedTraces)
	removeEmptyAttributesFromTraces(cleanedTraces)

	// Reuse envelope structures for performance - avoid per-span allocations
	tempTd := ptrace.NewTraces()
	tempRs := tempTd.ResourceSpans().AppendEmpty()
	tempSs := tempRs.ScopeSpans().AppendEmpty()

	rss := cleanedTraces.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		// Copy resource once per ResourceSpans
		rs.Resource().CopyTo(tempRs.Resource())
		tempRs.SetSchemaUrl(rs.SchemaUrl())

		sss := rs.ScopeSpans()
		for j := 0; j < sss.Len(); j++ {
			ss := sss.At(j)

			// Copy scope once per ScopeSpans
			ss.Scope().CopyTo(tempSs.Scope())
			tempSs.SetSchemaUrl(ss.SchemaUrl())

			spans := ss.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				// Clear previous span and add current one - reuse envelope
				tempSpans := tempSs.Spans()
				tempSpans.RemoveIf(func(ptrace.Span) bool { return true })
				span.CopyTo(tempSpans.AppendEmpty())

				jsonBytes, err := marshaler.MarshalTraces(tempTd)
				if err != nil {
					return nil, err
				}

				buf.Write(jsonBytes)
				buf.WriteByte('\n')
			}
		}
	}

	return buf.Bytes(), nil
}

