package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//at init create a servers loader

type servers struct {
	url   string
	count int
}

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
		var s = []servers{{
			url: "http://localhost:8080",
		}, {url: "http://localhost:8081"}, {url: "http://localhost:8082"}}
		// logic's
		proxy := Proxy(s[0].url)
		proxy.ServeHTTP(w, r)
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func Proxy(server string) *httputil.ReverseProxy {
	url, err := url.Parse(server)
	if err != nil {
		fmt.Println("errorr", err.Error())
		return nil
	}
	return httputil.NewSingleHostReverseProxy(url)
}
