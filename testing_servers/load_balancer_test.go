package testingservers

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_LoadBalancer(t *testing.T) {
	for range 100 {
		t.Run("API test", Test_api)
		time.Sleep(500 * time.Millisecond)
	}
}

// var current int

func Test_api(t *testing.T) {
	// var urls = []string{"/hello", "/api/math/add"}
	// current := rand.IntN(2)
	// + urls[current]
	res, err := http.Get("http://localhost:3001/api/math/add")
	if err != nil {
		fmt.Println("error", err.Error())
		return
	}
	defer res.Body.Close()
	reader, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("response error", err.Error())
		return
	}
	fmt.Println("status", res.StatusCode, "response", string(reader))
}
