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
	})
}
