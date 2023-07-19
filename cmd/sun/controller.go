package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Controller struct {
	activeConfig      Config
	in                <-chan Message
	publish           chan<- []byte
	updatePeriod      time.Duration
	localTime         time.Time
	overrideTime      time.Time
	usingOverrideTime bool
	grbl              *GrblArduino
}

func NewController(inChan <-chan Message, outChan chan<- []byte, grbl *GrblArduino) Controller {
	defaultPeriod, _ := time.ParseDuration("10s")
	defaultLat, defaultLong := degToRad(174.8860), degToRad(-40.9006)
	initialTime := time.Date(2023, 1, 1, 12, 00, 0, 0, time.Local)
	return Controller{
		activeConfig:      Config{Location: Location{Lat: defaultLat, Long: defaultLong}},
		in:                inChan,
		publish:           outChan,
		updatePeriod:      defaultPeriod,
		localTime:         initialTime, //TODO back to time.Now()
		overrideTime:      time.Now(),
		usingOverrideTime: false,
		grbl:              grbl,
	}
}

func (c Controller) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.updatePeriod) //this triggers an update to be sent via GCode //update frequency for the robot
	for {
		select {

		case <-ctx.Done():
			//disable steppers, shutdown active commands, we're going down...
			log.Printf("Terminating control loop, see ya.")
			return nil

		case msg := <-c.in:
			//apply update to config.. perhaps a JSON merge
			log.Printf("Recieved inwards message with type: %v", msg.T)

			switch msg.T {

			case "config":
				uc := Config{}
				json.Unmarshal(msg.D, &uc)
				c.HandleConfigUpdate(uc)

			case "move":
				mtr := MoveTargetRelative{}
				json.Unmarshal(msg.D, &mtr)
				c.HandleTargetAdjustment(mtr)

			default:
				log.Printf("Controller dropped message with type %v as no handler defined.", msg.T)
			}

		case tick := <-ticker.C:
			//recalculate desired position
			mAzi, mAlt := c.RecalculateDesiredMirrorPosition()
			log.Printf("DesiredMirrorPos: azi:%v, alt:%v", mAzi, mAlt)
			mAzi_Deg := radToDeg(mAzi)
			mAlt_Deg := radToDeg(mAlt)
			//convert position to GCode..
			code, err := PositionToGCode(mAzi_Deg, mAlt_Deg)
			if err != nil {
				log.Printf("%v", err)
			}
			log.Printf("Sending %v to grbl at: %v", string(code), tick)
			resp, err := c.grbl.GrblSendCommandGetResponse(code)
			if err != nil {
				log.Printf("err: %v", err)
			}
			log.Printf("Response was: %v", string(resp))
			c.publish <- []byte(fmt.Sprintf("Sent %v to grbl at: %v", string(code), tick))
		}
	}
}
