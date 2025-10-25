package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/AdarshN-1324/load_balancer-reverse_proxy/server_conf"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

var mode string

func main() {
	mode = os.Getenv("mode")

	run_mode := ""

	switch mode {
	case "WRR":
		run_mode = "Weighted Round robin"
	default:
		run_mode = "Round robin"
	}

	port := ":3001"
	log.Printf("Welcome to the simple Load balancer running in %s mode...\nListening and serving HTTP on %s\n", run_mode, port)

	serverpool := server_conf.Loadservers()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProxyRequestHandler(w, r, serverpool) // Pass the pool to the handler
	})

	log.Println(http.ListenAndServe(port, http.HandlerFunc(handler)))
}

func Handlerfunc(w http.ResponseWriter, r *http.Request) {

}

func ProxyRequestHandler(w http.ResponseWriter, r *http.Request, serverpool *server_conf.Server) {

	switch r.URL.Path {
	case "/ping":
		Ping(w, r)
	default:
		var current int
		switch mode {
		case "WRR":
			current = serverpool.WrrGetCurrent()
		default:
			current = serverpool.RRGetCurrent()
		}
		if current < 0 {
			log.Println("all the servers are in-active")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Bad Gateway Please try again"))
			return
		}
		r.Host = serverpool.Backends[current].Url.Host
		r.URL.Host = serverpool.Backends[current].Url.Host
		serverpool.Backends[current].Proxy.ServeHTTP(w, r)
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
