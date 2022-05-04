package main

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/config"
	"github.com/QuangTung97/promo-readonly/pkg/grpclib"
	"github.com/QuangTung97/promo-readonly/pkg/otellib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/go-sql-driver/mysql"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

//revive:disable-next-line:unused-parameter
func registerGRPCGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) {
}

func startServer() {
	conf := config.Load()
	logger := config.NewLogger(conf.Log)

	tracerProvider, shutdown := otellib.InitOtel("qrsearch-api", "local", conf.Jaeger)
	defer shutdown()

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(grpclib.RecoveryHandlerFunc)),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,

			otellib.UnaryServerInterceptor(tracerProvider),
			otellib.SetTraceInfoInterceptor(logger),

			grpc_zap.UnaryServerInterceptor(logger),
			grpc_zap.PayloadUnaryServerInterceptor(logger, payloadLogDecider),

			// Custom Interceptors Here
		),
		grpc.ChainStreamInterceptor(
			grpc_recovery.StreamServerInterceptor(),
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger),
		),
	)

	// db := conf.MySQL.MustConnect()

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)

	startHTTPAndGRPCServers(conf, grpcServer)
}

func main() {
	rootCmd := cobra.Command{
		Use: "server",
	}
	rootCmd.AddCommand(
		startServerCommand(),
	)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func payloadLogDecider(_ context.Context, _ string, _ interface{}) bool {
	return true
}

func startServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start the server",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}
}

func startHTTPAndGRPCServers(conf config.Config, grpcServer *grpc.Server) {
	fmt.Println("GRPC:", conf.Server.GRPC.ListenString())
	fmt.Println("HTTP:", conf.Server.HTTP.ListenString())

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	ctx := context.Background()
	grpcHost := conf.Server.GRPC.String()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	registerGRPCGateway(ctx, mux, grpcHost, opts)

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.Handle("/", mux)

	httpServer := &http.Server{
		Addr:    conf.Server.HTTP.ListenString(),
		Handler: httpMux,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
		fmt.Println("Shutdown HTTP server successfully")
	}()

	go func() {
		defer wg.Done()

		listener, err := net.Listen("tcp", conf.Server.GRPC.ListenString())
		if err != nil {
			panic(err)
		}

		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
		fmt.Println("Shutdown gRPC server successfully")
	}()

	//--------------------------------
	// Graceful Shutdown
	//--------------------------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	err := httpServer.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	wg.Wait()
}
