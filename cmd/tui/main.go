package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"

	sun "github.com/mykldog7/heliostat2/pkg/types"
	"github.com/rivo/tview"
	"nhooyr.io/websocket"
)

var (
	app             *tview.Application //main app
	details         *tview.Flex        //details (right-hand) pane
	notes           *tview.TextView    //place to provide text details to the operator
	actions         *tview.List        //actions (left-hand) pane
	selectedAction  string             //label of the currently selected action
	currentMoveSize float64            //size to adjust the target by in relative mode
	config          *sun.Config        //config structure/values returned from controller
	address         *string            //address/url of the websocket endpoint
	conn            *websocket.Conn    //connection to the server
	toServer        chan []byte        //channel to send messages to the server, ui will send messages here
)

func main() {
	currentMoveSize = math.Pi / 180 //a single degree

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	//Command line arguments (if any)
	address = flag.String("connect", "ws://localhost:8080", "provide a ws endpoint where controller is running")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Error channel
	errChan := make(chan error)

	//Make UI wait for connection
	var uiCanStart sync.Mutex
	uiCanStart.Lock() //lock the UI until we are connected to the server

	//Start Connection to Server
	go func(ctx context.Context, alllowUI *sync.Mutex) {
		err := StartConnection(ctx, *address, alllowUI)
		if err != nil {
			errChan <- fmt.Errorf("Error with connection to %v, Error: %v", *address, err)
		}
	}(ctx, &uiCanStart)

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
	select {
	case <-sigs:
		log.Printf("Recieved termination signal...")
		app.Stop()
		cancel()

	case err := <-errChan:
		if err.Error() != "UI Exited" { //was it a normal exit
			log.Printf("Recieved error: %v", err)
		}
		app.Stop()
		cancel()
	}

}
