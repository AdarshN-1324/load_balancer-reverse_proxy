package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdarshN-1324/load_balancer-reverse_proxy/server_conf"
)

func Checkconn() {
	for i := range server_conf.Servers.Urls {
		Checkactive(i)
	}
}

func Checkactive(i int) {
	defer server_conf.Servers.Mux.Unlock()
	server_conf.Servers.Mux.Lock()
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	url := server_conf.Servers.Urls[i].Url.String() + "/api/ping"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	cli := http.Client{}
	res, err := cli.Do(req)
	if err != nil {
		fmt.Println("response error:", err.Error())
		return
	}
	fmt.Println("host", url, "status code", res.StatusCode)
}
