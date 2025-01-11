package receiver

import (
	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	"github.com/jiekun/metrics-noop-receiver/zstd"
	writev2 "github.com/prometheus/prometheus/prompb/io/prometheus/write/v2"
	"io"
	"log"
	"strconv"
)

var (
	prometheusRemoteWriteV2RequestTotal              = metrics.NewCounter(`requests_total{path="/api/v2/write"}`)
	prometheusRemoteWriteV2ReadErrorTotal            = metrics.NewCounter(`read_error_total{path="/api/v2/write"}`)
	prometheusRemoteWriteV2DecodeErrorTotal          = metrics.NewCounter(`decode_error_total{path="/api/v2/write"}`)
	prometheusRemoteWriteV2PrometheusSampleTotal     = metrics.NewCounter(`sampled_total{path="/api/v2/write",exporter="prometheus-3"}`)
	prometheusRemoteWriteV2PrometheusZstdSampleTotal = metrics.NewCounter(`sampled_total{path="/api/v2/write",exporter="prometheus-3-zstd"}`)
)

func NewPrometheusRemoteWriteV2Route(r *gin.Engine) {
	r.POST("/api/v2/write", func(c *gin.Context) {
		sampleCnt, histCnt, ExemplarCnt := 0, 0, 0
		prometheusRemoteWriteV2RequestTotal.Inc()
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			prometheusRemoteWriteV2ReadErrorTotal.Inc()
			return
		}
		var body []byte
		contentEnc := c.Request.Header.Get("Content-Encoding")
		if contentEnc == "snappy" {
			body, err = snappy.Decode(body, b)
			if err != nil {
				log.Printf("snappy.Decode err: %v\n", err)
				prometheusRemoteWriteV2DecodeErrorTotal.Inc()
				return
			}
			defer func() {
				prometheusRemoteWriteV2PrometheusSampleTotal.Add(sampleCnt)
			}()
		} else if contentEnc == "zstd" {
			body, err = zstd.Decompress(body, b)
			if err != nil {
				log.Printf("zstd.Decompress err: %v\n", err)
				prometheusRemoteWriteV2DecodeErrorTotal.Inc()
				return
			}
			defer func() {
				prometheusRemoteWriteV2PrometheusZstdSampleTotal.Add(sampleCnt)
			}()
		} else {
			log.Printf("unsupported Content-Encoding: %v\n", contentEnc)
			prometheusRemoteWriteV2DecodeErrorTotal.Inc()
			return
		}

		request := &writev2.Request{}
		err = request.Unmarshal(body)
		if err != nil {
			log.Printf("json unmarshal write request err: %v\n", err)
			prometheusRemoteWriteV2DecodeErrorTotal.Inc()
			return
		}
		ts := request.GetTimeseries()
		for i := range ts {
			sampleCnt += len(ts[i].GetSamples())
			histCnt += len(ts[i].GetHistograms())
			ExemplarCnt += len(ts[i].GetExemplars())
		}
		c.Writer.Header().Set("X-Prometheus-Remote-Write-Samples-Written", strconv.Itoa(sampleCnt))
		c.Writer.Header().Set("X-Prometheus-Remote-Write-Histograms-Written", strconv.Itoa(histCnt))
		c.Writer.Header().Set("X-Prometheus-Remote-Write-Exemplars-Written", strconv.Itoa(ExemplarCnt))
		return
	})
}
