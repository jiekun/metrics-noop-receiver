package receiver

import (
	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
	otlp "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
)

var (
	otlpHTTPExportRequestTotal     = metrics.NewCounter(`requests_total{path="/otlp/export"}`)
	otlpHTTPExportReadErrorTotal   = metrics.NewCounter(`read_error_total{path="/otlp/export"}`)
	otlpHTTPExportDecodeErrorTotal = metrics.NewCounter(`decode_error_total{path="/otlp/export"}`)
	otlpHTTPExportSampleTotal      = metrics.NewCounter(`sampled_total{path="/otlp/export"}`)
)

func NewOTLPHTTPRoute(r *gin.Engine) {
	r.POST("/otlp/export", func(c *gin.Context) {
		otlpHTTPExportRequestTotal.Inc()
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			otlpHTTPExportReadErrorTotal.Inc()
			return
		}
		req := &otlp.ExportMetricsServiceRequest{}
		err = proto.Unmarshal(b, req)
		if err != nil {
			log.Printf("proto.Unmarshal err: %v\n", err)
			otlpHTTPExportDecodeErrorTotal.Inc()
			return
		}

		for _, rs := range req.GetResourceMetrics() {
			for _, sm := range rs.GetScopeMetrics() {
				for _, m := range sm.GetMetrics() {
					gauge := m.GetGauge()
					exponentialHistogram := m.GetExponentialHistogram()
					sum := m.GetSum()
					summary := m.GetSummary()
					otlpHTTPExportSampleTotal.Add(len(gauge.GetDataPoints()) + len(exponentialHistogram.GetDataPoints()) + len(sum.GetDataPoints()) + len(summary.GetDataPoints()))
				}
			}
		}
	})
}
