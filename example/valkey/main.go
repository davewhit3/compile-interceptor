package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davewhit3/compile-interceptor/dashboard"
	valkey "github.com/valkey-io/valkey-go"
)

const CacheKey = "vvvv-key"

func main() {
	fmt.Println("Hello, World Valkey!")
	done := make(chan os.Signal, 1)
	stopChan := make(chan struct{})
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go func() {
		reader := time.NewTicker(5 * time.Second)
		timer := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-timer.C:
				valkeyReq(client)
			case <-reader.C:
				readerReq(client)
			case <-stopChan:
				timer.Stop()
				reader.Stop()
				return
			}
		}
	}()

	mux := http.NewServeMux()
	dashboard.Register(mux)

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

func valkeyReq(client valkey.Client) {
	fmt.Println("Making request")
	err := client.Do(
		context.Background(),
		client.
			B().Set().
			Key(CacheKey).Value(time.Now().String()).
			ExSeconds(int64(30)).Build(),
	).Error()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func readerReq(client valkey.Client) {
	cmd := client.B().Get().Key(CacheKey).Build()
	result := client.Do(context.Background(), cmd)

	if err := result.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return
		}
	}

	value, err := result.ToString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("value", value)
}
