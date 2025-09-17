// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bytes"
	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func removeNullValues(data []byte) []byte {
	if !bytes.Contains(data, []byte(": null")) {
		return data
	}

	nullFieldRegex := regexp.MustCompile(`"[^"]+"\s*:\s*null\s*,?\s*`)
	result := nullFieldRegex.ReplaceAll(data, []byte(""))

	result = regexp.MustCompile(`,\s*}`).ReplaceAll(result, []byte("}"))
	result = regexp.MustCompile(`,\s*]`).ReplaceAll(result, []byte("]"))

	return result
}

// removeEmptyAttributesFromTraces efficiently removes empty attributes that cause Jaeger errors
// This prevents "invalid tag type in <nil>" by cleaning traces in-place once instead of per-span reconstruction
func removeEmptyAttributesFromTraces(td ptrace.Traces) {
	cleanAttrs := func(attrs pcommon.Map) {
		attrs.RemoveIf(func(_ string, v pcommon.Value) bool {
			return v.Type() == pcommon.ValueTypeEmpty
		})
	}

	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		cleanAttrs(rs.Resource().Attributes())

		sss := rs.ScopeSpans()
		for j := 0; j < sss.Len(); j++ {
			ss := sss.At(j)
			cleanAttrs(ss.Scope().Attributes())

			spans := ss.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				cleanAttrs(span.Attributes())

				// Clean span events
				events := span.Events()
				for l := 0; l < events.Len(); l++ {
					cleanAttrs(events.At(l).Attributes())
				}

				// Clean span links
				links := span.Links()
				for l := 0; l < links.Len(); l++ {
					cleanAttrs(links.At(l).Attributes())
				}
			}
		}
	}
}