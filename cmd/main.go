package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"dirbuster"
	"time"
)

func main() {

	baseurl := flag.String("u", "http://127.0.0.1", "Host to test")
	wordlist := flag.String("w", "wordlist.txt", "paths to test")

	flag.Parse()

	fmt.Println(Files("../"))

	list, err := dirbuster.ListToCheck(*wordlist)
	if err != nil {
		log.Fatal(err)
	}
	dirbuster.Exists(*baseurl, list)
}

func Files(path string) (count int) {
	fsys := os.DirFS(path)
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if filepath.Ext(p) == ".go" {
			count++
		}
		return nil
	})
	return count
}
