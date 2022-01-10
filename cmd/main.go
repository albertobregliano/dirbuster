package main

import (
	"context"
	"dirbuster"
	"flag"
	"log"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	baseurl := flag.String("u", "http://127.0.0.1", "Host to test")
	wordlist := flag.String("w", "wordlist.txt", "paths to test")
	output := flag.String("o", "", "output file to store data")

	flag.Parse()

	err := dirbuster.Run(ctx, *baseurl, *wordlist, *output)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

}
