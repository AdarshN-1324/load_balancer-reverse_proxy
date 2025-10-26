package server_conf

import (
	"context"
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
	Active         atomic.Bool
	// Mux            sync.Mutex
	weights int
}

type Server struct {
	Backends []Backend
	Current  int //to move for the next (round robin) /store the current index
	Mux      sync.Mutex
}

func Loadservers() *Server {

	var Servers Server
	servers := os.Getenv("servers")
	list := strings.Split(servers, ",")

	weights := os.Getenv("weights")
	w_list := strings.Split(weights, ",")

	Servers.Backends = make([]Backend, len(list))

	var wg sync.WaitGroup

	for i := range list {

		// Servers.Backends[i].Mux = sync.Mutex{}
		url, proxy := serverpool(list[i])

		Servers.Backends[i].Url, Servers.Backends[i].Proxy = url, proxy

		wg.Add(1)
		go Servers.Backends[i].CheckActive(&wg, &Servers)

		weight, err := strconv.ParseFloat(w_list[i], 64)
		if err != nil {
			log.Println("error", err.Error())
		}
		Servers.Backends[i].weights = int(weight * 10)
	}

	wg.Wait()

	return &Servers
	// Servers.total = len(Servers.Urls)
}

func serverpool(server_url string) (*url.URL, *httputil.ReverseProxy) {

	url, err := url.Parse(server_url)

	if err != nil {
		log.Println("error", err.Error())
		return nil, nil
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return url, proxy

}

func (backend *Backend) CheckActive(wg *sync.WaitGroup, serverpool *Server) {
	defer wg.Done()

	// defer backend.Mux.Unlock()
	// backend.Mux.Lock()

	path := backend.Url.String() + os.Getenv("active_path")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		log.Printf("Request create Error %s", err.Error())
		return
	}

	_, err = client.Do(req)
	if err != nil {
		log.Printf("Server Ping Error %s", err.Error())
		backend.Active.Store(false)
		return
	}

	backend.Active.Store(true)
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
a loop with a simple math solves the active backend issue
*/
func (server *Server) GetActiveBackend(current int) int {

	backend_active := server.Backends[server.Current].Active.Load()
	if backend_active {
		server.Backends[server.Current].Increment()
		return server.Current
	}

	for i := 0; i < len(server.Backends); i++ {
		idx := (server.Current + i) % len(server.Backends)
		backend_active = server.Backends[idx].Active.Load()
		if backend_active {
			server.Current = idx
			server.Backends[server.Current].Increment()
			return server.Current
		}
	}

	return -1
}

func (server *Server) RRGetCurrent() int {
	server.Mux.Lock()
	defer server.Mux.Unlock()

	server.Current++

	if server.Current >= len(server.Backends) {
		server.Current = 0
	}

	return server.GetActiveBackend(server.Current)
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

	return server.GetActiveBackend(server.Current)
}
