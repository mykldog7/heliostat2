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

	inwards := make(chan Message) //messages coming into the controller
	publish := make(chan []byte)  //messages to be pushed out to each subscriber

	ctx, cancel := context.WithCancel(context.Background())

	//Start our websocket server(provides interface to manage controller)
	go func() {
		err := startWebsocketServer("localhost:8080", ctx, inwards, publish)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//Controller is used to run the primary control loop, updating calculations and sending commands to grbl
	go func() {
		//Initialize and connect to the GRBL motor controller
		grbl, err := NewGrblArduino(ctx)
		if err != nil {
			log.Fatal(err)
		}
		Controller := NewController(inwards, publish, grbl)
		err = Controller.Start(ctx)
		if err != nil {
			log.Fatalf("Problem with controller %v", err)
		}
		log.Print("Controller shutdown completed")
	}()

	//wait for a termination signal, then clean-up when its recieved
	<-sigs
	log.Printf("Terminating...")
	cancel()
	time.Sleep(1 * time.Second)
}
