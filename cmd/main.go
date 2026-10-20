package main

import (
	"fmt"

	"github.com/Danikarat/double-entry-ledger/internal/config"
	"github.com/Danikarat/double-entry-ledger/internal/server"
)



func main(){

	fmt.Println("Hello double entry ledger")

	cfg := config.Load()
	srv := server.New(cfg, )
	
}