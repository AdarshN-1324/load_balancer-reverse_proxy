package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var count int

func Requests() {
	count++
	fmt.Println("requests", count)
}
func PING(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	w.Write([]byte("pong"))
}

func Add(w http.ResponseWriter, r *http.Request) {
	defer Requests()
	defer r.Body.Close()

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	var number struct {
		No uint `json:"no"`
	}
	json.Unmarshal(resp, &number)
	number.No += number.No
	w.Write([]byte(fmt.Sprint(number.No)))
}

func Sub(w http.ResponseWriter, r *http.Request) {
	defer Requests()
	defer r.Body.Close()

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	var number struct {
		No uint `json:"no"`
	}
	json.Unmarshal(resp, &number)
	number.No -= number.No
	w.Write([]byte(fmt.Sprint(number.No)))
}

func Pow(w http.ResponseWriter, r *http.Request) {
	defer Requests()
	defer r.Body.Close()

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	var number struct {
		No uint `json:"no"`
	}
	json.Unmarshal(resp, &number)
	number.No *= number.No
	w.Write([]byte(fmt.Sprint(number.No)))
}

func main() {
	port := "8080"
	fmt.Printf("welcome to simple math web listining under port %s \n", port)
	http.HandleFunc("/api/ping", PING)
	http.HandleFunc("/api/math/add", Add)
	http.HandleFunc("/api/math/sub", Sub)
	http.HandleFunc("/api/math/pow", Pow)
	http.ListenAndServe(":"+port, nil)

}
