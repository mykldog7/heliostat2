package main

import (
	"context"
	"log"
	"time"
)

type Controller struct {
	activeConfig      Config
	in                <-chan interface{}
	out               chan<- []byte
	updatePeriod      time.Duration
	localTime         time.Time
	overrideTime      time.Time
	usingOverrideTime bool
}

func NewController(inChan <-chan interface{}, outChan chan<- []byte) Controller {
	defaultPeriod, _ := time.ParseDuration("30s")
	return Controller{
		activeConfig:      Config{},
		in:                inChan,
		out:               outChan,
		updatePeriod:      defaultPeriod,
		localTime:         time.Now(),
		overrideTime:      time.Now(),
		usingOverrideTime: false,
	}
}

func (c Controller) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.updatePeriod) //this triggers an update to be sent via GCode //update frequency for the robot
	for {
		select {
		case <-ctx.Done():
			//disable steppers, shutdown active commands, we're going down...
			log.Printf("disabling steppers, disconnecting from GCode layer, shutdown control loop, see ya.")
			return nil
		case signal := <-c.in:
			//apply update to config.. perhaps a JSON merge
			log.Printf("Recieved inwards object: %v", signal)
		case tick := <-ticker.C:
			//apply update to robot? tick holds the time
			log.Printf("stub move robot at: %v", tick)
		}
	}
}
