package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/sixdouglas/suncalc"
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
	defaultPeriod, _ := time.ParseDuration("30s")
	return Controller{
		activeConfig:      Config{},
		in:                inChan,
		publish:           outChan,
		updatePeriod:      defaultPeriod,
		localTime:         time.Now(),
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
			tAzi, tAlt := c.RecalculateTarget()
			tAzi_Deg := radToDeg(tAzi)
			tAlt_Deg := radToDeg(tAlt)
			//convert position to GCode..
			code := PositionToGCode(tAzi_Deg, tAlt_Deg)
			//apply update to robot? tick holds the time
			c.grbl.toGrbl <- code
			log.Printf("Sent %v to grbl at: %v", string(code), tick)
			c.publish <- []byte(fmt.Sprintf("Sent %v to grbl at: %v", string(code), tick))

		case msg := <-c.grbl.fromGrbl:
			log.Printf("Grbl says: %v", string(msg))

		}
	}
}

// RecalculateTarget returns the position correct according to configured time
// Altitude: sun altitude above the horizon in radians, e.g. 0 at the horizon and PI/2 at the zenith (straight over your head)
// Azimuth: sun azimuth in radians (direction along the horizon, measured from south to west), e.g. 0 is south and Math.PI * 3/4 is northwest
func (c *Controller) RecalculateTarget() (float64, float64) {
	var instant time.Time
	if c.usingOverrideTime {
		instant = c.overrideTime
	} else {
		instant = c.localTime
	}
	sun := suncalc.GetPosition(instant, c.activeConfig.Location.Lat, c.activeConfig.Location.Long)
	mirrorAzi, mirrorAlt := calculateMirrorTarget(sun.Azimuth, sun.Altitude, c.activeConfig.Target.Azimuth, c.activeConfig.Target.Altitude)
	return mirrorAzi, mirrorAlt
}

// HandleConfigUpdate updates the 'activeconfig on the controller
func (c *Controller) HandleConfigUpdate(uc Config) {
}

func (c *Controller) HandleTargetAdjustment(m MoveTargetRelative) {
	switch m.Direction {
	case "up":
		c.activeConfig.Target.Altitude += m.Amount
		if c.activeConfig.Target.Altitude > (math.Pi / 4.0) {
			c.activeConfig.Target.Altitude = math.Pi / 4.0 //cap at vertical
		}
	case "down":
		c.activeConfig.Target.Altitude -= m.Amount
		if c.activeConfig.Target.Altitude < 0 {
			c.activeConfig.Target.Altitude = 0 //cap at horizon
		}
	case "left":
		c.activeConfig.Target.Azimuth -= m.Amount
		if c.activeConfig.Target.Azimuth < 0 {
			c.activeConfig.Target.Azimuth += (math.Pi * 2) //wrap around the circle
		}
	case "right":
		c.activeConfig.Target.Azimuth += m.Amount
		if c.activeConfig.Target.Azimuth > (math.Pi * 2) {
			c.activeConfig.Target.Azimuth -= (math.Pi * 2) //wrap around the circle
		}
	}
}
