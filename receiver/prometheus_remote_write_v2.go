package receiver

import (
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
		c.Request.ParseForm()
		if err != nil {
			prometheusRemoteWriteReadErrorTotal.Inc()
			return
		}

		var body []byte
		body, err = snappy.Decode(body, b)
		if err != nil {
			log.Printf("snappy.Decode err: %v\n", err)
			prometheusRemoteWriteDecodeErrorTotal.Inc()
			return
		}

		writeRequest := &prompb.WriteRequest{}
		err = writeRequest.Unmarshal(body)
		if err != nil {
			log.Printf("json unmarshal write request err: %v\n", err)
			prometheusRemoteWriteDecodeErrorTotal.Inc()
			return
		}
	})
}
