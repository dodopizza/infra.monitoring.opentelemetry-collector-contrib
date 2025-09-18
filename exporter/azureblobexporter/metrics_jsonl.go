// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bytes"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type metricsJSONLMarshaler struct{}

func (*metricsJSONLMarshaler) MarshalMetrics(md pmetric.Metrics) ([]byte, error) {
	marshaler := &pmetric.JSONMarshaler{}
	var buf bytes.Buffer

	tempMd := pmetric.NewMetrics()
	tempRm := tempMd.ResourceMetrics().AppendEmpty()
	tempSm := tempRm.ScopeMetrics().AppendEmpty()
	tempMetrics := tempSm.Metrics()
	tempMetrics.EnsureCapacity(1)
	if tempMetrics.Len() == 0 {
		tempMetrics.AppendEmpty()
	}

	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)

		rm.Resource().CopyTo(tempRm.Resource())
		cleanAttrs(tempRm.Resource().Attributes())
		tempRm.SetSchemaUrl(rm.SchemaUrl())

		sms := rm.ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			sm := sms.At(j)

			sm.Scope().CopyTo(tempSm.Scope())
			cleanAttrs(tempSm.Scope().Attributes())
			tempSm.SetSchemaUrl(sm.SchemaUrl())

			metrics := sm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)

				metric.CopyTo(tempMetrics.At(0))
				tempMetric := tempMetrics.At(0)

				cleanMetricDataPoints(tempMetric)

				jsonBytes, err := marshaler.MarshalMetrics(tempMd)
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
func cleanMetricDataPoints(metric pmetric.Metric) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		dps := metric.Gauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			cleanAttrs(dps.At(i).Attributes())
		}
	case pmetric.MetricTypeSum:
		dps := metric.Sum().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			cleanAttrs(dps.At(i).Attributes())
		}
	case pmetric.MetricTypeHistogram:
		dps := metric.Histogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			cleanAttrs(dps.At(i).Attributes())
		}
	case pmetric.MetricTypeExponentialHistogram:
		dps := metric.ExponentialHistogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			cleanAttrs(dps.At(i).Attributes())
		}
	case pmetric.MetricTypeSummary:
		dps := metric.Summary().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			cleanAttrs(dps.At(i).Attributes())
		}
	}
}