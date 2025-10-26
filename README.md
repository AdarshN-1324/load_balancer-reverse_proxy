# Go Load Balancer & Reverse Proxy

A simple, lightweight, and extensible HTTP load balancer and reverse proxy written in Go. This project demonstrates core load balancing concepts including different distribution strategies, backend health checks, and graceful shutdowns.

## Features

-   **Multiple Balancing Strategies**:
    -   **Round Robin (RR)**: Distributes requests sequentially across the pool of backend servers.
    -   **Weighted Round Robin (WRR)**: Distributes requests based on a predefined weight for each server. Servers with higher weights receive more requests.
-   **Backend Health Checks**: A concurrent worker periodically pings backend servers on a configurable health endpoint to ensure they are active. Inactive servers are automatically removed from and added back to the rotation.
-   **Graceful Shutdown**: On receiving an interrupt signal (`SIGINT`), the server will stop accepting new connections and wait for active requests to complete before shutting down.
-   **Configuration via Environment Variables**: Easily configure backend servers, weights, and balancing mode using a `.env` file.
-   **High Concurrency**: Built using Go's powerful concurrency primitives (`goroutines`, `atomic` operations) for efficient, non-blocking operations.

## Getting Started

Follow these instructions to get a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

-   Go (version 1.19 or newer recommended for `atomic.Bool`)

### Installation

1.  Clone the repository:
    ```sh
    git clone <your-repository-url>
    cd load_balancer-reverse_proxy
    ```

2.  Install dependencies:
    ```sh
    go mod tidy
    ```

### Configuration

The application is configured using environment variables. Create a `.env` file in the root of the project. You can copy the example below.

**`.env.example`**
```env
# A comma-separated list of backend server URLs
servers=http://localhost:8081,http://localhost:8082,http://localhost:8083

# The load balancing mode. Options: "RR" (Round Robin) or "WRR" (Weighted Round Robin)
mode=WRR

# A comma-separated list of weights corresponding to the servers list.
# Used only in WRR mode. The values are relative (e.g.,if there's two servers(server1,server2)  0.8 is 8 requests for server1 and 0.2 is 2 requests for the server2).
weights=0.3,0.2,0.5

# The path used by the health checker to ping backend servers.
# The backend servers should return a 200 OK on this path.
active_path=/health
```

### Running the Application

1.  Ensure you have some backend servers running that respond to the `active_path`. For example, a simple Go server:
    ```go
    // simple_backend/main.go
    package main
    import "net/http"
    func main() {
        http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
        http.ListenAndServe(":8081", nil)
    }
    ```

2.  Start the load balancer:
    ```sh
    go run main.go
    ```

You should see the following output:
```
Welcome to the simple Load balancer running in Weighted Round robin mode...
Listening and serving HTTP on :3001
```

## Usage

Send HTTP requests to the load balancer on port `3001`. The requests will be forwarded to your backend servers according to the configured balancing strategy.

```sh
# Example using curl
curl http://localhost:3001/some/path
```

The load balancer also exposes a `/ping` endpoint for checking its own status:
```sh
curl http://localhost:3001/ping
# Expected output: PONG
```

## Project Structure

```
├── go.mod
├── main.go               # Entrypoint, HTTP server setup, and graceful shutdown logic.
├── server_conf/
│   └── backends.go       # Core logic for managing the server pool, balancing algorithms, and health checks.
├── worker/
│   └── worker.go         # A concurrent worker that periodically runs health checks on all backends.
└── todo.txt              # Project goals and status.
```

## Future Improvements

-   **Structured Logging**: Implement structured logging (e.g., using `slog` or `zerolog`) to make logs machine-readable and easier to parse, which fulfills the final item on the `todo.txt`.
-   **Configuration from File**: Add support for loading configuration from a YAML or TOML file in addition to environment variables.
-   **More Algorithms**: Implement other load balancing algorithms like Least Connections.

