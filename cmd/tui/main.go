package main

import (
	"context"
	"flag"
	"log"
	"math"
	"os"
	"os/signal"

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
	connected       bool               //set if we are connected to a server
	conn            *websocket.Conn    //connection to the server
	ctx             context.Context    //context for the entire application, allows us to close the connection if anywhere in the app decides we are closing
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

	go func() {
		err := StartConnection(ctx, *address)
		if err != nil {
			log.Fatalf("Error with connection to %v. Error:%v", *address, err)
		}
	}()

	go func() {
		err := StartInterface()
		if err != nil {
			log.Fatalf("UI Err:%v", err)
		}
	}()

	//wait for a termination signal, then clean-up when its recieved
	<-sigs
	log.Printf("Terminating...")
	cancel()
}
