package main

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/config"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/cacheclient"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/pkg/memtable"
	"github.com/QuangTung97/promo-readonly/pkg/util"
	"github.com/QuangTung97/promo-readonly/promopb"
	"github.com/QuangTung97/promo-readonly/repository"
	"github.com/QuangTung97/promo-readonly/service/readonly"
	"github.com/spf13/cobra"
	"sort"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	rootCmd := cobra.Command{
		Use: "bench",
	}
	rootCmd.AddCommand(
		benchWithMemcachedCommand(),
		migrateDataCommand(),
	)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func benchWithMemcached() {
	conf := config.Load()
	fmt.Println("DBONLY:", conf.DBOnly)

	numConns := 1
	if conf.Memcache.NumConns > 0 {
		numConns = conf.Memcache.NumConns
	}
	fmt.Println("NUM CONNS:", numConns)

	fmt.Println("MEMCACHE ADDR:", conf.Memcache.Addr())

	db := conf.MySQL.MustConnect()
	provider := repository.NewProvider(db)
	client := cacheclient.New(conf.Memcache.Addr(), numConns)
	memTable := memtable.New(8 * 1024 * 1024)
	dhashProvider := dhash.NewProvider(memTable, client)

	const numThreads = 50
	const numElements = 2000

	durations := make([][]time.Duration, numThreads)

	server := readonly.NewServer(provider, dhashProvider, conf.DBOnly)

	totalStart := time.Now()

	var wg sync.WaitGroup
	wg.Add(numThreads)
	for th := 0; th < numThreads; th++ {
		threadIndex := th
		go func() {
			defer wg.Done()

			for i := 0; i < numElements; i++ {
				start := time.Now()
				resp, err := server.Check(context.Background(), &promopb.PromoServiceCheckRequest{
					Inputs: []*promopb.PromoServiceCheckInput{
						{
							VoucherCode:  "VOUCHER01",
							MerchantCode: "MERCHANT01",
							TerminalCode: "TERMINAL01",
							Phone:        "0987000111",
						},
						{
							VoucherCode:  "VOUCHER02",
							MerchantCode: "MERCHANT02",
							TerminalCode: "TERMINAL01",
							Phone:        "0987000222",
						},
					},
				})
				if err != nil {
					fmt.Println(resp, err)
				}
				durations[threadIndex] = append(durations[threadIndex], time.Since(start))
			}
		}()
	}
	wg.Wait()
	fmt.Println("TOTAL TIME", time.Since(totalStart))

	history := make([]time.Duration, 0, numThreads*numElements)

	total := time.Duration(0)
	for _, bucket := range durations {
		for _, d := range bucket {
			total += d
			history = append(history, d)
		}
	}
	avg := float64(total/time.Second) / float64(numThreads*numElements)

	sort.Slice(history, func(i, j int) bool {
		return history[i] < history[j]
	})

	numHistory := numElements * numThreads
	p50Index := numHistory * 50 / 100
	p90Index := numHistory * 90 / 100
	p95Index := numHistory * 95 / 100
	p99Index := numHistory * 99 / 100
	p999Index := numHistory * 999 / 1000

	fmt.Println("P50:", history[p50Index])
	fmt.Println("P90:", history[p90Index])
	fmt.Println("P95:", history[p95Index])
	fmt.Println("P99:", history[p99Index])
	fmt.Println("P999:", history[p999Index])
	fmt.Println("MAX:", history[numHistory-1])
	fmt.Println("HISTORY LEN:", len(history))

	fmt.Println("AVG:", avg)
}

func benchWithMemcachedCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "memcache",
		Short: "benchmark with memcached",
		Run: func(cmd *cobra.Command, args []string) {
			benchWithMemcached()
		},
	}
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
