package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {

	var wg sync.WaitGroup

	for index := 0; index < 10; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			fmt.Println("Index", index, time.Now())
			result, err := http.Get("http://localhost:8080/")
			if err != nil {
				fmt.Println("Error when doing request:", err)
				return
			}
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}
			fmt.Println("Response Body:", string(body))
		}(index)
	}

	wg.Wait()
}
