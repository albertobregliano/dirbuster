package dirbuster

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
)

func Exist(url string) (*http.Response, error) {

	return http.Head(url)
}

func ListToCheck(wordlist string) ([]string, error) {

	readFile, err := os.Open(wordlist)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var result []string

	for fileScanner.Scan() {
		result = append(result, fileScanner.Text())
	}

	sort.Slice(result, func(i, j int) bool { return result[i] > result[j] })

	return result, nil
}

func Exists(baseurl string, wordlist []string) {

	if !strings.HasSuffix(baseurl, "/") {
		baseurl += "/"
	}

	c := http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxIdleConns:           10,
			MaxIdleConnsPerHost:    10,
			MaxConnsPerHost:        10,
			IdleConnTimeout:        0,
			ResponseHeaderTimeout:  0,
			ExpectContinueTimeout:  0,
			MaxResponseHeaderBytes: 0,
			WriteBufferSize:        0,
			ReadBufferSize:         0,
			ForceAttemptHTTP2:      false,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 0,
	}

	var wg sync.WaitGroup

	var results = make(chan string)
	var finished = make(chan bool, 1)

	go func() {
		for {
			r, ok := <-results
			if !ok {
				break
			}
			fmt.Println(r)
		}
		finished <- true
	}()

	for _, word := range wordlist {
		uri := baseurl + word
		_, err := url.ParseRequestURI(uri)
		if err != nil {
			log.Printf("%s is not a valid url: %v\n", uri, err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := c.Head(uri)
			if err != nil {
				log.Printf(uri, err)
			}
			if resp.StatusCode <= 403 {
				results <- uri + " " + resp.Status
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()
	close(results)
	<-finished
}
