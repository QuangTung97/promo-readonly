package main

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/config"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/cacheclient"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/pkg/grpclib"
	"github.com/QuangTung97/promo-readonly/pkg/memtable"
	"github.com/QuangTung97/promo-readonly/pkg/otellib"
	"github.com/QuangTung97/promo-readonly/pkg/util"
	"github.com/QuangTung97/promo-readonly/promopb"
	"github.com/QuangTung97/promo-readonly/repository"
	"github.com/QuangTung97/promo-readonly/service/readonly"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/credentials/insecure"
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
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

//revive:disable-next-line:unused-parameter
func registerGRPCGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) {
	_ = promopb.RegisterPromoServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

func startServer() {
	conf := config.Load()
	logger := config.NewLogger(conf.Log)

	tracerProvider, shutdown := otellib.InitOtel("promo-api", "local", conf.Jaeger)
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

	memTable := memtable.New(16 * 1024 * 1024)
	client := cacheclient.New("localhost:11211", 4)

	db := conf.MySQL.MustConnect()
	provider := repository.NewProvider(db)
	dhashProvider := dhash.NewProvider(memTable, client)

	promoServer := readonly.NewServer(provider, dhashProvider, conf.DBOnly)
	promopb.RegisterPromoServiceServer(grpcServer, promoServer)

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
		migrateDataCommand(),
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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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

func migrateDataCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "migrate data",
		Run: func(cmd *cobra.Command, args []string) {
			conf := config.Load()
			db := conf.MySQL.MustConnect()

			provider := repository.NewProvider(db)
			repo := repository.NewBlacklist()
			err := provider.Transact(context.Background(), func(ctx context.Context) error {
				err := repo.UpsertConfig(ctx, model.BlacklistConfig{
					CustomerCount: 1,
					MerchantCount: 1,
					TerminalCount: 0,
				})
				if err != nil {
					return err
				}

				err = repo.UpsertBlacklistMerchants(ctx, []model.BlacklistMerchant{
					{
						Hash:         util.HashFunc("MERCHANT01"),
						MerchantCode: "MERCHANT01",
						Status:       model.BlacklistMerchantStatusActive,
					},
				})
				if err != nil {
					return err
				}

				err = repo.UpsertBlacklistCustomers(ctx, []model.BlacklistCustomer{
					{
						Hash:   util.HashFunc("0987000111"),
						Phone:  "0987000111",
						Status: model.BlacklistCustomerStatusActive,
					},
				})
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				panic(err)
			}
		},
	}
}
