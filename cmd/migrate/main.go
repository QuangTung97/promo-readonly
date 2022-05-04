package main

import (
	"fmt"
	"github.com/QuangTung97/promo-readonly/config"
	"github.com/QuangTung97/promo-readonly/pkg/migration"
)

func main() {
	conf := config.Load()
	cmd := migration.MigrateCommand(conf.MySQL.DSN())
	err := cmd.Execute()
	if err != nil {
		fmt.Println("[ERROR]", err)
	}
}
