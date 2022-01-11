package main

import (
	"context"
	"dirbuster"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	defer func() {
		log.Println("Stop received")
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	baseurl := flag.String("u", "http://127.0.0.1", "Host to test")
	wordlist := flag.String("w", "wordlist.txt", "paths to test")
	output := flag.String("o", "", "output file to store data")

	flag.Parse()

	err := dirbuster.Run(ctx, *baseurl, *wordlist, *output)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

}
