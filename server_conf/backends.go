package server_conf

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log"
)

type Backend struct {
	Url            *url.URL
	Proxy          *httputil.ReverseProxy
	Requests       atomic.Uint64 // if using WRR reset to 0 when requests meet count
	Total_Requests atomic.Uint64
	Active         bool
	Mux            sync.Mutex
	weights        int
}

type Server struct {
	Backends       []Backend
	Inactive_count int
	Current        int //to move for the next (round robin) /store the current index
	// total   int //round robin is a bit difficult when you look for active one
	Mux sync.Mutex
}

func Loadservers() *Server {
	var Servers Server
	servers := os.Getenv("servers")
	list := strings.Split(servers, ",")

	weights := os.Getenv("weights")
	w_list := strings.Split(weights, ",")

	Servers.Backends = make([]Backend, len(list))

	for i := range list {

		Servers.Backends[i].Mux = sync.Mutex{}
		url, proxy := serverpool(list[i])

		Servers.Backends[i].Url, Servers.Backends[i].Proxy = url, proxy

		go Servers.Backends[i].CheckActive()

		weight, err := strconv.ParseFloat(w_list[i], 64)
		if err != nil {
			fmt.Println("error", err.Error())
		}
		Servers.Backends[i].weights = int(weight * 10)
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

func (backend *Backend) CheckActive() {
	path := backend.Url.String() + os.Getenv("active_path")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		log.Printf("Request create Error %s", err.Error())
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("Server Ping Error %s", err.Error())
		return
	}

	backend.Active = res.StatusCode == 200
}

func (backend *Backend) Increment() {
	backend.Requests.Add(1)
	backend.Total_Requests.Add(1)
}

// func (server *Server) IsActive()bool{
// 	return server.
// }

/*
write a recursive function to get a current active one in case a server is down  and i need to keep on checking if the next one is in active then increment
*/
func (server *Server) GetActiveBackend(current int) int {

	if server.Inactive_count == len(server.Backends)-1 {
		return -1
	}
	if server.Backends[current].Active {
		server.Backends[server.Current].Increment()
		return current
	}

	server.Inactive_count++

	if current >= len(server.Backends) {
		current = 0
	}
	return server.GetActiveBackend(current + 1)
}

func (server *Server) RRGetCurrent() int {
	server.Mux.Lock()
	defer server.Mux.Unlock()

	server.Current++

	if server.Current >= len(server.Backends) {
		server.Current = 0
	}
	server.Current = server.GetActiveBackend(server.Current)

	return server.Current
}

func (server *Server) WrrGetCurrent() int {
	server.Mux.Lock()
	defer server.Mux.Unlock()

	requests := server.Backends[server.Current].Requests.Load()
	if int(requests) >= server.Backends[server.Current].weights {
		server.Backends[server.Current].Requests.Store(0)
		server.Current++
		if server.Current >= len(server.Backends) {
			server.Current = 0
		}
	}

	// once incremented look for active/inactive and then icrement again

	return server.GetActiveBackend(server.Current)
}
