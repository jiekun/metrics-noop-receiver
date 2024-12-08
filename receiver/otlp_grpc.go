package receiver

import (
	"context"
	"github.com/VictoriaMetrics/metrics"
	otlp "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
)

var (
	otlpExportRequestTotal     = metrics.NewCounter(`requests_total{path="otlp.export"}`)
	otlpExportDecodeErrorTotal = metrics.NewCounter(`decode_error_total{path="otlp.export"}`)
)

type noopOTLPMetricsServer struct {
	otlp.UnimplementedMetricsServiceServer
}

func NewOTLPMetricsEndpoint(grpcServer *grpc.Server) {
	otlp.RegisterMetricsServiceServer(grpcServer, &noopOTLPMetricsServer{})
}

func (s *noopOTLPMetricsServer) Export(context.Context, *otlp.ExportMetricsServiceRequest) (*otlp.ExportMetricsServiceResponse, error) {
	otlpExportRequestTotal.Inc()
	return &otlp.ExportMetricsServiceResponse{}, nil
}
