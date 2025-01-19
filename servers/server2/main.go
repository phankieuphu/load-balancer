package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, server 2 received : %s\n", r.URL.Path)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server at port 8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		fmt.Println(err)
	}
}
