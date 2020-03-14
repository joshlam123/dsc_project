package main

import (
	"fmt"
	"net/http"
	"time"
)

var counter int = 0
var ch chan int = make(chan int)

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 10)
	fmt.Println(r.URL.Path, r.Body)
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":3000", nil)
}
