package main

import (
	"context"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	otlp "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

var (
	prometheusRemoteWriteRequestTotal     = metrics.NewCounter(`requests_total{path="/api/v1/write"}`)
	prometheusRemoteWriteReadErrorTotal   = metrics.NewCounter(`read_error_total{path="/api/v1/write"}`)
	prometheusRemoteWriteDecodeErrorTotal = metrics.NewCounter(`decode_error_total{path="/api/v1/write"}`)
)

func main() {
	initHTTPServer()
	initGRPCServer()
}

func initHTTPServer() {
	r := gin.Default()
	{
		r.POST("/api/v1/write", func(c *gin.Context) {
			prometheusRemoteWriteRequestTotal.Inc()
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				prometheusRemoteWriteReadErrorTotal.Inc()
			}
			var body []byte
			_, err = snappy.Decode(b, body)
			if err != nil {
				prometheusRemoteWriteDecodeErrorTotal.Inc()
			}
		})
	}
	go r.Run()
}

func initGRPCServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8001))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	otlp.RegisterMetricsServiceServer(grpcServer, &noopOTLPMetricsServer{})
	go grpcServer.Serve(lis)
}

type noopOTLPMetricsServer struct {
	otlp.UnimplementedMetricsServiceServer
}

func (s *noopOTLPMetricsServer) Export(context.Context, *otlp.ExportMetricsServiceRequest) (*otlp.ExportMetricsServiceResponse, error) {
	return &otlp.ExportMetricsServiceResponse{}, nil
}
