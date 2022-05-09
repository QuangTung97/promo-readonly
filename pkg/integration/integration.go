package integration

import (
	"fmt"
	"github.com/QuangTung97/promo-readonly/config"
	"github.com/QuangTung97/promo-readonly/pkg/migration"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"os"
	"path"
	"sync"

	// for integration test, must not be imported in any main.go
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// TestCase ...
type TestCase struct {
	DB   *sqlx.DB
	Conf config.Config
}

var initOnce sync.Once

var globalConf config.Config
var globalDB *sqlx.DB

// NewTestCase ...
func NewTestCase() *TestCase {
	initOnce.Do(func() {
		rootDir := findRootDir()

		conf := config.LoadTestConfig(rootDir)
		migration.MigrateUpForTesting(rootDir, conf.MySQL.DSN())

		db := conf.MySQL.MustConnect()

		globalConf = conf
		globalDB = db
	})

	return &TestCase{
		Conf: globalConf,
		DB:   globalDB,
	}
}

// Truncate ...
func (tc *TestCase) Truncate(db string) {
	tc.DB.MustExec(fmt.Sprintf("TRUNCATE %s", db))
}

func findRootDir() string {
	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	directory := workdir
	for {
		files, err := ioutil.ReadDir(directory)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if file.Name() == "go.mod" {
				return directory
			}
		}

		directory = path.Dir(directory)
	}
}
