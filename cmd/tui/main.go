package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mykldog7/heliostat2/pkg/types"
	sun "github.com/mykldog7/heliostat2/pkg/types"
)

var (
	selectedAction  string             //label of the currently selected action
	currentMoveSize float64            //size to adjust the target by in relative mode
	config          *sun.Config        //config structure/values returned from controller
	address         *string            //address/url of the websocket endpoint
	toServer        chan types.Message //channel of messages that we will send to server
)

func main() {
	currentMoveSize = math.Pi / 180 //a single degree

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	//Command line arguments (if any)
	address = flag.String("connect", "ws://localhost:8080", "provide a ws endpoint where controller is running")
	flag.Parse()

	//Error channel
	errChan := make(chan error)

	//Make UI wait for connection
	var uiCanStart sync.Mutex
	uiCanStart.Lock() //lock the UI until we are connected to the server

	// Outgoing Channel
	toServer = make(chan types.Message)

	//Start Connection to Server
	go func(alllowUI *sync.Mutex) {
		err := StartConnection(*address, alllowUI)
		if err != nil {
			errChan <- fmt.Errorf("Error with connection to %v, Error: %v", *address, err)
		}
	}(&uiCanStart)

	//START UI
	go func(mu *sync.Mutex) {
		uiCanStart.Lock() //wait for connection to be established, before starting the UI
		err := StartInterface()
		if err != nil {
			errChan <- fmt.Errorf("Error with UI, Error: %v", err)
		}
		errChan <- fmt.Errorf("UI Exited") //this is a normal exit
	}(&uiCanStart)

	//wait for a termination signal, or error, then clean-up when its recieved
	for {
		select {
		case <-sigs:
			errChan <- fmt.Errorf("Termination Signal")

		case err := <-errChan:
			if err.Error() != "UI Exited" { //was it a normal exit
				log.Printf("Recieved: %v", err)
				os.Exit(1)
			}
			toServer <- sun.Message{T: "close"}
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		}
	}
}
