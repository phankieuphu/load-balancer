package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	fmt.Fprintf(w, "Hello, server 1 received: %s\n", r.URL.Path)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server at port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println(err)
	}
}
