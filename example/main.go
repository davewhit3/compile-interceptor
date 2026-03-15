package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Hello, World!")

	resp, err := http.Get("https://www.google.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
}
