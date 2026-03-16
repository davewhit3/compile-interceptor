package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davewhit3/compile-interceptor/outgoing"
)

var ctr = 0

func main() {
	fmt.Println("Hello, World!")
	done := make(chan os.Signal)
	stopChan := make(chan struct{})
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		timer := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-timer.C:
				ctr++
				fmt.Println("ctr", ctr)
				httpReq()
			case <-stopChan:
				timer.Stop()
			}
		}
	}()

	http.HandleFunc("/", outgoingRequest)

	go func() {
		fmt.Println("Starting server")
		err := http.ListenAndServe(":8080", nil)
		fmt.Println("Server stopped")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}()

	fmt.Println("Waiting for signal")
	<-done
	fmt.Println("Signal received")
	stopChan <- struct{}{}
	fmt.Println("Stopping server")
}

func outgoingRequest(w http.ResponseWriter, r *http.Request) {
	for _, url := range outgoing.List() {
		fmt.Fprintf(w, "%s\n", url)
	}
	fmt.Fprintf(w, "------\n")
}

func httpReq() {
	fmt.Println("Making request")
	resp, err := http.Get(fmt.Sprintf("https://tvn24.pl?zzz=%d", time.Now().UnixNano()))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
}
