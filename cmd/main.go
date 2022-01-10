package main

import (
	"context"
	"dirbuster"
	"flag"
	"log"
	"os"
)

func main() {

	ctx := context.Background()

	baseurl := flag.String("u", "http://127.0.0.1", "Host to test")
	wordlist := flag.String("w", "wordlist.txt", "paths to test")
	output := flag.String("o", "", "output file to store data")

	flag.Parse()

	b, err := dirbuster.NewBuster(
		dirbuster.WithBaseurl(*baseurl),
		dirbuster.WithWordlist(*wordlist),
		dirbuster.WithContext(ctx),
	)
	if err != nil {
		log.Fatalf("impossible to create buster, error: %v", err)
	}

	if *output != "" {
		outputFile, err := os.Create(*output)
		if err != nil {
			log.Fatalf("impossible to create file %s, error: %v", *output, err)
		}
		setOutput := dirbuster.WithOutput(outputFile)
		setOutput(&b)
	}

	b.Run()
}
