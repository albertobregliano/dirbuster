package dirbuster

import (
	"bufio"
	"context"
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
	"time"
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

var results = make(chan string, 1)
var tobeanalized = make(chan string)
var tobetested = make(chan string)
var finished = make(chan bool, 1)
var mapTestedUri = make(map[string]bool)
var sem = make(chan int, 4)

var c http.Client

func Exists(ctx context.Context, b buster) error {

	words, err := ListToCheck(b.wordlist)
	if err != nil {
		return fmt.Errorf("wordlist is not valid: %v", err)
	}

	baseurlpieces, err := url.Parse(b.baseurl)
	if err != nil {
		return fmt.Errorf("baseurl is not valid: %v", err)
	}
	baseurlScheme := baseurlpieces.Scheme
	// baseurlPort := baseurlpieces.Port()
	baseurlHost := baseurlpieces.Host

	if !strings.HasSuffix(b.baseurl, "/") {
		b.baseurl += "/"
	}

	customT := &http.Transport{
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
	}

	c = http.Client{
		Transport: retryRoundTripper{
			next:       customT,
			maxRetries: 3,
			delay:      1 * time.Second,
		},
		Timeout: 0,
	}

	var wg sync.WaitGroup

	go func() {
		for {
			uri, ok := <-tobetested
			if !ok {
				break
			}
			if !mapTestedUri[uri] {
				mapTestedUri[uri] = true
				wg.Add(1)
				go headPage(ctx, uri, &wg)
			}
		}
	}()

	go func() {
		for {
			uri, ok := <-tobeanalized
			if !ok {
				break
			}
			wg.Add(1)
			go getPage(ctx, uri, baseurlHost, baseurlScheme, &wg)
		}
	}()

	go func() {
		for {
			r, ok := <-results
			if !ok {
				break
			}
			b.output.Write([]byte(r + "\n"))
			//fmt.Println(r)
		}
		finished <- true
	}()
	for _, word := range words {
		uri := b.baseurl + word
		_, err := url.ParseRequestURI(uri)
		if err != nil {
			log.Printf("%s is not a valid url: %v\n", uri, err)
			continue
		}
		mapTestedUri[uri] = true
		wg.Add(1)
		go headPage(ctx, uri, &wg)
	}
	wg.Wait()
	log.Println("all go routines finished")
	close(tobeanalized)
	close(tobetested)
	close(results)
	<-finished
	return nil
}

func headPage(ctx context.Context, uri string, wg *sync.WaitGroup) {
	defer wg.Done()
	sem <- 1
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, uri, nil)
	if err != nil {
		log.Printf("Error in request creation: %v", err)
		return
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Printf(uri, err)
		return
	}
	if resp.StatusCode <= 403 {
		results <- fmt.Sprintf("%s\t%s\t%s\t%s", resp.Status, resp.Header.Get("Content-type"), resp.Header.Get("Content-length"), uri)
		if strings.Contains(resp.Header.Get("Content-type"), "html") {
			tobeanalized <- uri
		}
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	<-sem
}

func getPage(ctx context.Context, uri, baseurlHost, baseurlScheme string, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		log.Printf("Request non vaid %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	links := getLinks(resp.Body)
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	for _, link := range links {
		urlpieces, err := url.Parse(link)
		if err != nil {
			continue
		}
		if urlpieces.Host == "" {
			urlpieces.Host = baseurlHost
		}
		if urlpieces.Host != baseurlHost {
			continue
		}

		tobetested <- fmt.Sprintf("%s://%s/%s", baseurlScheme, baseurlHost, urlpieces.Path)
	}
}

type retryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
	delay      time.Duration
}

func (rr retryRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var attempts int
	for {
		res, err := rr.next.RoundTrip(r)
		attempts++

		// max attempts reached.
		if attempts == rr.maxRetries {
			return res, err
		}

		// good enough to stop retrying.
		if err == nil && res.StatusCode < http.StatusInternalServerError {
			return res, err
		}

		// delay and retry
		select {
		case <-r.Context().Done():
			return res, r.Context().Err()
		case <-time.After(rr.delay):
			log.Printf("will retry %s %s, %d of %d attempts\n", r.Method, r.URL, attempts, rr.maxRetries)
		}
	}
}
