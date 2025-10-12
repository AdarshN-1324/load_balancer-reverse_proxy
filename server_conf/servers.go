package server_conf

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"

	"log"
)

type Url struct {
	Url      *url.URL
	Proxy    *httputil.ReverseProxy
	Requests int
	Active   bool
	Mux      sync.Mutex
}

type Server struct {
	Urls    []Url
	Current int //to move for the next (round robin) /store the current index
	total   int //round robin is a bit difficult when you look for active one
}

var Servers Server

func Loadservers() {
	servers := os.Getenv("servers")
	list := strings.Split(servers, ",")
	Servers.Urls = make([]Url, len(list))
	for i := range list {
		fmt.Println(list[i])
		Servers.Urls[i].Mux = sync.Mutex{}
		url, proxy := CreateSerUrl(list[i])
		Servers.Urls[i].Url, Servers.Urls[i].Proxy = url, proxy
		Servers.Urls[i].Active = Servers.Urls[i].CheckActive()
	}
	Servers.total = len(Servers.Urls)
}
func CreateSerUrl(server_url string) (*url.URL, *httputil.ReverseProxy) {
	url, err := url.Parse(server_url)
	if err != nil {
		fmt.Println("errorr", err.Error())
		return nil, nil
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	return url, proxy

}
func (url *Url) CheckActive() bool {
	path := url.Url.String() + os.Getenv("active_path")
	res, err := http.Get(path)
	if err != nil {
		log.Printf("Server Ping Error %s", err.Error())
		return false
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

func (url *Url) Increment() {
	defer url.Mux.Unlock()
	url.Mux.Lock()
	url.Requests++
}

func (server *Server) GetCurrent() int {

	server.Current++
	if server.Current > server.total-1 {
		server.Current = 0
	}
	server.Urls[Servers.Current].Requests++
	return server.Current
}
