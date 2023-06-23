package main

import (
	"log"
	"time"
)

func Control(send chan<- interface{}, update_config <-chan Config, terminate <-chan bool) error {
	ticker := time.NewTicker(5 * time.Second) //this triggers an update to be sent via GCode //update frequency for the robot

	for {
		select {
		case update := <-update_config:
			//apply update to config.. perhaps a JSON merge
			log.Printf("Recieved config update object: %v", update)
		case tick := <-ticker.C:
			//apply update to robot? tick holds the time
			log.Printf("stub move robot at: %v", tick)
		case <-terminate:
			log.Printf("Terminating Control Loop")
			return nil
		}
	}

}
