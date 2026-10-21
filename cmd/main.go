package main

import (
	"fmt"
	"log"

	"github.com/dani-dzotsi/double-entry-ledger/internal/config"
	"github.com/dani-dzotsi/double-entry-ledger/internal/server"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	fmt.Println("Hello double entry ledger")

	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil{
		log.Fatal(err)
	}

	srv := server.New(cfg,db)
	if err := srv.Run(); err != nil{
		log.Fatal(err)
	}

}
