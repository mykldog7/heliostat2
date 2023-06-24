package main

import (
	"context"
	"log"
	"os"
	"os/signal"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	log.SetFlags(0)

	inwards := make(chan interface{})
	outwards := make(chan []byte)
	go func() {
		outwards <- []byte("welcome")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := startWebsocketServer("localhost:8080", ctx, inwards, outwards)
		if err != nil {
			log.Fatal(err)
		}
	}()
	//later this will be replaced by the control loop
	go func() {
		for {
			log.Printf("in:%v", <-inwards)
		}
	}()
	//wait for a termination signal
	<-sigs
	log.Printf("terminating")
}
