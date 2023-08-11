package main

import (
	"encoding/json"
	"log"
	"math"

	"github.com/mykldog7/heliostat2/pkg/types"
)

// HandleConfigUpdate updates the 'activeconfig on the controller
func (c *Controller) HandleGetActiveConfig() {
	log.Printf("GetActiveConfig")
	payload, err := json.Marshal(c.activeConfig)
	msg := types.Message{T: "ActiveConfig", D: payload}
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Print(err)
	}
	c.publish <- bytes
}

// HandleConfigUpdate updates the 'activeconfig on the controller
func (c *Controller) HandleConfigUpdate(uc types.Config) {
	log.Printf("UpdateConfig")
	c.activeConfig.Location.Lat = uc.Location.Lat
	c.activeConfig.Location.Long = uc.Location.Long
}

func (c *Controller) HandleTargetAdjustment(m types.MoveTargetRelative) {
	log.Printf("TargetAdjustment")
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
	default:
		c.publish <- types.NewAckMessage(false)
	}
	//if we got one of the expected directions send an ack
	if m.Direction == "up" || m.Direction == "down" || m.Direction == "left" || m.Direction == "right" {
		c.publish <- types.NewAckMessage(true)
	}
}
