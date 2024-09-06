package main

import (
	v11 "flix-indexer/flix/v1_1"
	"flix-indexer/server"
	"log"
	"os"
	"strconv"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	indexer := v11.NewIndexer()
	err := indexer.SeedFromFs()
	if err != nil {
		panic(err)
	}

	port := 8080
	envPort := os.Getenv("PORT")
	if envPort != "" {
		port, err = strconv.Atoi(envPort)
		if err != nil {
			panic(err)
		}
	}

	httpServer := server.HttpServer{
		Logger:  logger,
		Indexer: indexer,
		Port:    port,
	}
	httpServer.Setup()
	logger.Printf("starting server on port %d\n", httpServer.Port)

	err = httpServer.Start()
	if err != nil {
		panic(err)
	}
}
