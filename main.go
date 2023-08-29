package main

import (
	"eth-balance-proxy/client"
	"eth-balance-proxy/router"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	rpcClient := client.NewRpcClient(c, os.Args[1])
	rt, err := router.NewRouter(rpcClient)
	if err != nil {
		log.Fatalf("error initializing router: %v", err)
	}
	fmt.Println("Service is ready...")

	if err != http.ListenAndServe(":1903", rt) {
		log.Fatalf("error initializing router: %v", err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit
}
