package main

import (
	"fmt"
	"github.com/QuangTung97/promo-readonly/config"
	"github.com/QuangTung97/promo-readonly/pkg/migration"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	conf := config.Load()
	cmd := migration.MigrateCommand(conf.MySQL.DSN())
	err := cmd.Execute()
	if err != nil {
		fmt.Println("[ERROR]", err)
	}
}
