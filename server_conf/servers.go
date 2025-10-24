package server_conf

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"log"
)

type Url struct {
	Backend  *url.URL
	Proxy    *httputil.ReverseProxy
	Requests atomic.Uint64 // if using WRR reset to 0 when requests meet count
	Active   bool
	Mux      sync.Mutex
	weights  int
}

type Server struct {
	Urls    []Url
	Current int //to move for the next (round robin) /store the current index
	// total   int //round robin is a bit difficult when you look for active one
	Mux sync.Mutex
}

func Loadservers() *Server {
	var Servers Server
	servers := os.Getenv("servers")
	list := strings.Split(servers, ",")
	weights := os.Getenv("weights")
	w_list := strings.Split(weights, ",")
	Servers.Urls = make([]Url, len(list))
	for i := range list {
		Servers.Urls[i].Mux = sync.Mutex{}
		url, proxy := serverpool(list[i])
		Servers.Urls[i].Backend, Servers.Urls[i].Proxy = url, proxy
		Servers.Urls[i].Active = Servers.Urls[i].CheckActive()
		weight, err := strconv.ParseFloat(w_list[i], 64)
		if err != nil {
			fmt.Println("error", err.Error())
		}
		Servers.Urls[i].weights = int(weight * 10)
	}
	return &Servers
	// Servers.total = len(Servers.Urls)
}
func serverpool(server_url string) (*url.URL, *httputil.ReverseProxy) {
	url, err := url.Parse(server_url)
	if err != nil {
		fmt.Println("errorr", err.Error())
		return nil, nil
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	return url, proxy

}
func (url *Url) CheckActive() bool {
	path := url.Backend.String() + os.Getenv("active_path")
	res, err := http.Get(path)
	if err != nil {
		log.Printf("Server Ping Error %s", err.Error())
		return false
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

func (url *Url) Increment() {
	url.Requests.Add(1)

}

func (server *Server) RRGetCurrent() int {
	server.Mux.Lock()
	defer server.Mux.Unlock()
	server.Current++
	if server.Current >= len(server.Urls) {
		server.Current = 0
	}
	server.Urls[server.Current].Increment()
	return server.Current
}

func (server *Server) WrrGetCurrent() int {
	server.Mux.Lock()
	defer server.Mux.Unlock()
	requests := server.Urls[server.Current].Requests.Load()
	if int(requests) >= server.Urls[server.Current].weights {
		server.Urls[server.Current].Requests.Store(0)
		server.Current++
		if server.Current >= len(server.Urls) {
			server.Current = 0
		}
	}
	server.Urls[server.Current].Increment()
	return server.Current
}
