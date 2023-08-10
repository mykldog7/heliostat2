package main

import (
	"context"
	"flag"
	"log"
	"math"

	sun "github.com/mykldog7/heliostat2/pkg/types"
	"github.com/rivo/tview"
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
	client          WSClient           //connection/client of the heliostat server
	connected       bool               //set if we are connected to a server
)

func main() {
	currentMoveSize = math.Pi / 180 //a single degree

	//Command line arguments (if any)
	address = flag.String("connect", "ws://localhost:8080", "provide a ws endpoint where controller is running")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := NewConnection(ctx, *address)
	if err != nil {
		log.Fatalf("Can't connect to %v. Error:%v", *address, err)
	}

	config, err = client.getConfig()
	if err != nil {
		log.Fatalf("Error getting config:%v", err)
	}

	err = runUI()
	if err != nil {
		log.Fatal("UI Err:%v", err)
	}
}
