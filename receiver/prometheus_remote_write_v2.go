package receiver

import (
	"encoding/json"
	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	prompb "github.com/prometheus/prometheus/prompb"
	"io"
	"log"
)

var (
	prometheusRemoteWriteRequestTotal     = metrics.NewCounter(`requests_total{path="/api/v1/write"}`)
	prometheusRemoteWriteReadErrorTotal   = metrics.NewCounter(`read_error_total{path="/api/v1/write"}`)
	prometheusRemoteWriteDecodeErrorTotal = metrics.NewCounter(`decode_error_total{path="/api/v1/write"}`)
)

func NewPrometheusRemoteWriteV2Route(r *gin.Engine) {
	r.POST("/api/v1/write", func(c *gin.Context) {
		prometheusRemoteWriteRequestTotal.Inc()
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			prometheusRemoteWriteReadErrorTotal.Inc()
			return
		}

		var body []byte
		b, err = snappy.Decode(b, body)
		if err != nil {
			log.Printf("snappy.Decode err: %v\n", err)
			prometheusRemoteWriteDecodeErrorTotal.Inc()
			return
		}

		writeRequest := prompb.WriteRequest{}
		err = json.Unmarshal(b, &writeRequest)
		if err != nil {
			log.Printf("json unmarshal write request err: %v\n", err)
			prometheusRemoteWriteDecodeErrorTotal.Inc()
			return
		}
	})
}
