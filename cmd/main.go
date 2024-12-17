package main

import (
	"context"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
	"github.com/jiekun/metrics-noop-receiver/receiver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	httpSvr := initHTTPServer()
	grpcSvr := initGRPCServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := httpSvr.Shutdown(ctx); err != nil {
		log.Fatal("http server shutdown:", err)
	}
	// catching ctx.Done(). timeout of 1 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("http server exited")

	grpcSvr.GracefulStop()
	log.Println("grpc server exited")
}

func initHTTPServer() *http.Server {
	r := gin.New()
	r.Use(gin.Recovery())

	// init route
	receiver.NewPrometheusRemoteWriteV1Route(r)
	receiver.NewPrometheusRemoteWriteV2Route(r)
	receiver.NewOTLPHTTPRoute(r)

	// init metrics endpoint
	r.GET("/metrics", func(c *gin.Context) {
		metrics.WritePrometheus(c.Writer, true)
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: r.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	return srv
}

func initGRPCServer() *grpc.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 8001))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// init endpoints
	receiver.NewOTLPMetricsEndpoint(grpcServer)

	go grpcServer.Serve(lis)

	return grpcServer
}
