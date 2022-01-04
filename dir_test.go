package dirbuster_test

import (
	"dirbuster"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var port, address string
var srv *httptest.Server

func init() {
	rand.Seed(time.Now().UnixNano())
	port = ":" + strconv.Itoa(8000+rand.Intn(1000))
	address = "127.0.0.1"
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

	srv = httptest.NewUnstartedServer(mux)
	srv.Listener = l
}
func TestExist(t *testing.T) {
	t.Parallel()
	srv.Start()
	defer srv.CloseClientConnections()
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
		{"1,", args{url: "http://" + address + port + "/test"}, &http.Response{
			StatusCode: 200,
		}, false},
		{"2,", args{url: "http://" + address + port + "/nop"}, &http.Response{
			StatusCode: 404,
		}, false},
		{"3,", args{url: "http://" + address + port + "/resource"}, &http.Response{
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
