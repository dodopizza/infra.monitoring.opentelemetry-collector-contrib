package azureblobexporter

import (
	"bytes"

	"go.opentelemetry.io/collector/pdata/plog"
)

type logsJSONLMarshaler struct{}

func (m *logsJSONLMarshaler) MarshalLogs(ld plog.Logs) ([]byte, error) {
	marshaler := &plog.JSONMarshaler{}
	var buf bytes.Buffer

	tempLd := plog.NewLogs()
	tempRl := tempLd.ResourceLogs().AppendEmpty()
	tempSl := tempRl.ScopeLogs().AppendEmpty()
	tempLrs := tempSl.LogRecords()
	tempLrs.EnsureCapacity(1)
	if tempLrs.Len() == 0 {
		tempLrs.AppendEmpty()
	}

	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)

		rl.Resource().CopyTo(tempRl.Resource())
		cleanAttrs(tempRl.Resource().Attributes())
		tempRl.SetSchemaUrl(rl.SchemaUrl())

		sls := rl.ScopeLogs()
		for j := 0; j < sls.Len(); j++ {
			sl := sls.At(j)

			sl.Scope().CopyTo(tempSl.Scope())
			cleanAttrs(tempSl.Scope().Attributes())
			tempSl.SetSchemaUrl(sl.SchemaUrl())

			lrs := sl.LogRecords()
			for k := 0; k < lrs.Len(); k++ {
				lr := lrs.At(k)

				lr.CopyTo(tempLrs.At(0))
				tempLr := tempLrs.At(0)

				cleanAttrs(tempLr.Attributes())

				jsonBytes, err := marshaler.MarshalLogs(tempLd)
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
