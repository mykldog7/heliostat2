package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	log.SetFlags(0)

	inwards := make(chan interface{})
	publish := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := startWebsocketServer("localhost:8080", ctx, inwards, publish)
		if err != nil {
			log.Fatal(err)
		}
	}()
	//later this will be replaced by the control loop
	go func() {
		Controller := NewController(inwards, publish)
		err := Controller.Start(ctx)
		if err != nil {
			log.Fatalf("Problem with controller %v", err)
		}
		log.Print("Controller shutdown completed")
	}()
	//wait for a termination signal, then clean-up when its recieved
	<-sigs
	cancel()
	log.Printf("terminating")
	time.Sleep(2 * time.Second)
}
