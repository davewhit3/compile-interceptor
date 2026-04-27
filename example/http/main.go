package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davewhit3/compile-interceptor/dashboard"
)

var ctr = 0

func main() {
	fmt.Println("Hello, World!")
	done := make(chan os.Signal, 1)
	stopChan := make(chan struct{})
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		timer := time.NewTicker(1 * time.Second)
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

	mux := http.NewServeMux()
	dashboard.Register(dashboard.ForMux(mux))

	go func() {
		fmt.Println("Starting server on :8080 — open http://localhost:8080/telescope")
		err := http.ListenAndServe(":8080", mux)
		if err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	fmt.Println("Waiting for signal")
	<-done
	fmt.Println("Signal received")
	stopChan <- struct{}{}
	fmt.Println("Stopping server")
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
