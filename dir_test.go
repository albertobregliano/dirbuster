package dirbuster_test

import (
	"dirbuster"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var portint int

func webserver() *httptest.Server {
	portint++
	port := ":" + strconv.Itoa(24000+portint)
	address := "127.0.0.1"
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
	srv.Listener = l
	return srv
}

func TestExist(t *testing.T) {
	t.Parallel()
	srv := webserver()
	srv.Start()
	defer srv.Close()

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Response
		wantErr bool
	}{
		{"1,", args{url: srv.URL + "/test"}, &http.Response{
			StatusCode: 200,
		}, false},
		{"2,", args{url: srv.URL + "/nop"}, &http.Response{
			StatusCode: 404,
		}, false},
		{"3,", args{url: srv.URL + "/resource"}, &http.Response{
			StatusCode: 403,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dirbuster.Exist(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.StatusCode != tt.want.StatusCode {
				t.Errorf("Exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExists(t *testing.T) {
	t.Parallel()
	srv := webserver()
	srv.Start()
	defer srv.CloseClientConnections()
	defer srv.Close()
	type args struct {
		baseurl  string
		wordlist []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"1", args{srv.URL, []string{"test"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirbuster.Exists(tt.args.baseurl, tt.args.wordlist)
		})
	}
	srv.Close()
}
