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
	"math/rand"
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

func randInt() int {
	return int(rand.Int63n(1000000))
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

				const inputSize = 20
				inputs := make([]*promopb.PromoServiceCheckInput, 0, inputSize)
				for k := 0; k < inputSize; k++ {
					inputs = append(inputs, &promopb.PromoServiceCheckInput{
						VoucherCode:  "VOUCHER01",
						MerchantCode: getMerchantCode(randInt()),
						TerminalCode: "TERMINAL01",
						Phone:        getPhone(randInt()),
					})
				}

				resp, err := server.Check(context.Background(), &promopb.PromoServiceCheckRequest{
					Inputs: inputs,
				})
				if err != nil {
					fmt.Println(resp, err)
				}
				durations[threadIndex] = append(durations[threadIndex], time.Since(start))
			}
		}()
	}
	wg.Wait()
	totalDuration := time.Since(totalStart)
	fmt.Println("TOTAL TIME", totalDuration)

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
	fmt.Println("QPS:", float64(numHistory)/totalDuration.Seconds())

	fmt.Println("ACCESS COUNT:", client.AccessCount())
	fmt.Println("MISS COUNT:", client.MissCount())
	fmt.Println("MISS RATE:", float64(client.MissCount())/float64(client.AccessCount())*100)
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

func getPhone(index int) string {
	return fmt.Sprintf("0987%06d", index)
}

func getMerchantCode(index int) string {
	return fmt.Sprintf("MERCHANT%06d", index)
}

const batchSize = 1000
const numBatch = 100

func migrateMerchants(ctx context.Context, repo repository.Blacklist) {
	for i := 0; i < numBatch; i++ {
		merchants := make([]model.BlacklistMerchant, 0, batchSize)
		for k := 0; k < batchSize; k++ {
			code := getMerchantCode(k + i*batchSize)
			merchants = append(merchants, model.BlacklistMerchant{
				Hash:         util.HashFunc(code),
				MerchantCode: code,
				Status:       model.BlacklistMerchantStatusActive,
			})
		}
		err := repo.UpsertBlacklistMerchants(ctx, merchants)
		if err != nil {
			panic(err)
		}
	}
}

func migrateCustomers(ctx context.Context, repo repository.Blacklist) {
	for i := 0; i < numBatch; i++ {
		customers := make([]model.BlacklistCustomer, 0, batchSize)
		for k := 0; k < batchSize; k++ {
			phone := getPhone(k + i*batchSize)
			customers = append(customers, model.BlacklistCustomer{
				Hash:   util.HashFunc(phone),
				Phone:  phone,
				Status: model.BlacklistCustomerStatusActive,
			})
		}
		err := repo.UpsertBlacklistCustomers(ctx, customers)
		if err != nil {
			panic(err)
		}
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
					CustomerCount: numBatch * batchSize,
					MerchantCount: numBatch * batchSize,
					TerminalCount: 0,
				})
				if err != nil {
					return err
				}

				migrateMerchants(ctx, repo)
				migrateCustomers(ctx, repo)

				return nil
			})
			if err != nil {
				panic(err)
			}
		},
	}
}
