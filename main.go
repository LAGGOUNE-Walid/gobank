package main

import (
	"log"

	"github.com/LAGGOUNE-Walid/gobank/api"
	"github.com/LAGGOUNE-Walid/gobank/storage"
)

func main() {
	store, err := storage.NewSqliteStore()
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewServer("0.0.0.0:8080", store)
	server.Run()
}
