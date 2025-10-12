package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/AdarshN-1324/load_balancer-reverse_proxy/server_conf"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	server_conf.Loadservers()
}

//at init create a servers loader

func main() {
	port := ":3001"
	fmt.Printf("Welcome to the simple Load balancer...\nListening and serving HTTP on %s\n", port)
	log.Print(http.ListenAndServe(port, http.HandlerFunc(ProxyRequestHandler)))
}

// this is where the main process is done passing the url to the server
func ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	// write this server paths and forward them as needed
	switch r.URL.Path {
	case "/ping":
		Ping(w, r)
	default:
		// logic part round robin
		current := server_conf.Servers.GetCurrent()
		r.Host = server_conf.Servers.Urls[current].Url.Host
		r.URL.Host = server_conf.Servers.Urls[current].Url.Host

		fmt.Println("current", current, "requests", server_conf.Servers.Urls[current].Requests)
		server_conf.Servers.Urls[current].Proxy.ServeHTTP(w, r)
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func Proxy(url *url.URL) *httputil.ReverseProxy {
	// url, _ := url.Parse(s_url)
	// fmt.Println("active status", server.Active)
	// fmt.Println("url", server.Url.String())
	return httputil.NewSingleHostReverseProxy(url)
}
