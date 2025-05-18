package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	testURL := "http://example.com"

	serverURL := "http://localhost:8080/shorten"

	for i := 1; i < 10; i++ {
		go func(iteration int) {
			body := fmt.Sprintf(`{"url": "%s"}`, testURL)
			response, err := http.Post(serverURL, "application/json", bytes.NewBuffer([]byte(body)))

			if err != nil {
				fmt.Printf("Request %d: Error: %v\n", iteration, err)
			}
			defer response.Body.Close()

			responseBody, _ := io.ReadAll(response.Body)
			fmt.Printf("Request %d: Response: %s\n", iteration, string(responseBody))
		}(i)
	}

	// Wait to let goroutines finish
	time.Sleep(10 * time.Second)
}
