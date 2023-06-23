package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	log.SetFlags(0)

	fromWebsocket := make(chan interface{})
	toWebsocket := make(chan interface{})

	go func() {
		err := startWebsocketServer("localhost:8080", fromWebsocket, toWebsocket)
		if err != nil {
			log.Fatal(err)
		}
	}()
	for {
		select {
		case sig := <-sigs:
			log.Printf("terminating: %v", sig)
		}
	}
}
