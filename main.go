package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AdarshN-1324/load_balancer-reverse_proxy/server_conf"
	"github.com/AdarshN-1324/load_balancer-reverse_proxy/worker"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func main() {

	mode := os.Getenv("mode")

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

	server := http.Server{
		Addr: port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProxyRequestHandler(w, r, serverpool, mode) // Pass the pool to the handler

		}),
	}

	var wg sync.WaitGroup

	// graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT)
	wg.Add(1)
	go func() {

		defer wg.Done()
		<-shutdown

		Timeout := 5 * time.Second
		log.Printf("Received shutdown signal. Commencing graceful shutdown with a %s timeout...", Timeout)

		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		// Attempt to gracefully shut down the server
		log.Println("Waiting for active requests to complete...")
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("server forced to shutdown with error %v", err)
		}

		log.Println("Server shutdown successful.")

	}()

	go worker.CheckBackend(serverpool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	wg.Wait()
}

func ProxyRequestHandler(w http.ResponseWriter, r *http.Request, serverpool *server_conf.Server, mode string) {

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

// func Proxy(url *url.URL) *httputil.ReverseProxy {
// 	// url, _ := url.Parse(s_url)
// 	// fmt.Println("active status", server.Active)
// 	// fmt.Println("url", server.Url.String())
// 	return httputil.NewSingleHostReverseProxy(url)
// }
