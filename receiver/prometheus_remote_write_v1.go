package receiver

import (
	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	"github.com/jiekun/metrics-noop-receiver/zstd"
	prompb "github.com/prometheus/prometheus/prompb"
	"io"
	"log"
)

var (
	prometheusRemoteWriteV1RequestTotal          = metrics.NewCounter(`requests_total{path="/api/v1/write"}`)
	prometheusRemoteWriteV1ReadErrorTotal        = metrics.NewCounter(`read_error_total{path="/api/v1/write"}`)
	prometheusRemoteWriteV1DecodeErrorTotal      = metrics.NewCounter(`decode_error_total{path="/api/v1/write"}`)
	prometheusRemoteWriteV1PrometheusSampleTotal = metrics.NewCounter(`sampled_total{path="/api/v1/write",exporter="prometheus-2"}`)
	prometheusRemoteWriteV1VMAgentSampleTotal    = metrics.NewCounter(`sampled_total{path="/api/v1/write",exporter="vmagent"}`)
)

func NewPrometheusRemoteWriteV1Route(r *gin.Engine) {
	r.POST("/api/v1/write", func(c *gin.Context) {
		sampleCnt, histCnt, ExemplarCnt := 0, 0, 0
		prometheusRemoteWriteV1RequestTotal.Inc()
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			prometheusRemoteWriteV1ReadErrorTotal.Inc()
			return
		}

		var body []byte
		if c.GetHeader("Content-Encoding") == "zstd" {
			body, err = zstd.Decompress(body, b)
			defer func() {
				prometheusRemoteWriteV1VMAgentSampleTotal.Add(sampleCnt)
			}()
		} else {
			body, err = snappy.Decode(body, b)
			defer func() {
				prometheusRemoteWriteV1PrometheusSampleTotal.Add(sampleCnt)
			}()
		}

		if err != nil {
			log.Printf("snappy.Decode err: %v\n", err)
			prometheusRemoteWriteV1DecodeErrorTotal.Inc()
			return
		}

		writeRequest := &prompb.WriteRequest{}
		err = writeRequest.Unmarshal(body)
		if err != nil {
			log.Printf("json unmarshal write request err: %v\n", err)
			prometheusRemoteWriteV1DecodeErrorTotal.Inc()
			return
		}

		ts := writeRequest.GetTimeseries()
		for i := range ts {
			sampleCnt += len(ts[i].GetSamples())
			histCnt += len(ts[i].GetHistograms())
			ExemplarCnt += len(ts[i].GetExemplars())
		}
	})
}
