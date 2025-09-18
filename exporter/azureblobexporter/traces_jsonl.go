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

	tempTd := ptrace.NewTraces()
	tempRs := tempTd.ResourceSpans().AppendEmpty()
	tempSs := tempRs.ScopeSpans().AppendEmpty()
	tempSpans := tempSs.Spans()
	tempSpans.EnsureCapacity(1)
	if tempSpans.Len() == 0 {
		tempSpans.AppendEmpty()
	}

	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		rs.Resource().CopyTo(tempRs.Resource())
		cleanAttrs(tempRs.Resource().Attributes())
		tempRs.SetSchemaUrl(rs.SchemaUrl())

		sss := rs.ScopeSpans()
		for j := 0; j < sss.Len(); j++ {
			ss := sss.At(j)

			ss.Scope().CopyTo(tempSs.Scope())
			cleanAttrs(tempSs.Scope().Attributes())
			tempSs.SetSchemaUrl(ss.SchemaUrl())

			spans := ss.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				span.CopyTo(tempSpans.At(0))
				tempSpan := tempSpans.At(0)

				cleanAttrs(tempSpan.Attributes())
				events := tempSpan.Events()
				for l := 0; l < events.Len(); l++ {
					cleanAttrs(events.At(l).Attributes())
				}
				links := tempSpan.Links()
				for l := 0; l < links.Len(); l++ {
					cleanAttrs(links.At(l).Attributes())
				}

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

