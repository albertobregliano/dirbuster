package dirbuster_test

import (
	"context"
	"dirbuster"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var portint int

func webserver() *httptest.Server {
	portint++
	address := "127.0.0.1"
	port := ":" + strconv.Itoa(2600+portint)

	l, err := net.Listen("tcp", address+port)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})
	mux.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	srv := httptest.NewUnstartedServer(mux)
	srv.Close()
	srv.Listener = l

	return srv
}

func TestRun(t *testing.T) {
	t.Parallel()

	srv := webserver()
	srv.Start()
	defer srv.Close()

	file, err := ioutil.TempFile("", "wordlist")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("/\ntest\n")

	o, err := ioutil.TempFile("", "output")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(o.Name())

	type args struct {
		ctx      context.Context
		baseurl  string
		wordlist string
		output   interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{context.TODO(), srv.URL, file.Name(), o}, false},
		{"2", args{context.TODO(), srv.URL, file.Name(), o.Name()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dirbuster.Run(tt.args.ctx, tt.args.baseurl, tt.args.wordlist, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
