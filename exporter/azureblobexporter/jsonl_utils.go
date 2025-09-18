// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bytes"
	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

var (
	nullFieldRegex             = regexp.MustCompile(`"[^"]+"\s*:\s*null\s*,?\s*`)
	trailingCommaInObjectRegex = regexp.MustCompile(`,\s*}`)
	trailingCommaInArrayRegex  = regexp.MustCompile(`,\s*]`)
)

func removeNullValues(data []byte) []byte {
	if !bytes.Contains(data, []byte(": null")) {
		return data
	}

	result := nullFieldRegex.ReplaceAll(data, []byte(""))
	result = trailingCommaInObjectRegex.ReplaceAll(result, []byte("}"))
	result = trailingCommaInArrayRegex.ReplaceAll(result, []byte("]"))

	return result
}

func cleanAttrs(attrs pcommon.Map) {
	attrs.RemoveIf(func(_ string, v pcommon.Value) bool {
		return v.Type() == pcommon.ValueTypeEmpty
	})
}
