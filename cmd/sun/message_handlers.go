package main

import (
	"encoding/json"
	"log"
	"math"

	sun "github.com/mykldog7/heliostat2/pkg/types"
)

// HandleConfigUpdate updates the 'activeconfig on the controller
func (c *Controller) HandleGetActiveConfig() {
	log.Printf("GetActiveConfig")
	payload, err := json.Marshal(c.activeConfig)
	msg := sun.Message{T: "ActiveConfig", D: payload}
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Print(err)
	}
	c.publish <- bytes
}

// HandleConfigUpdate updates the 'activeconfig on the controller
func (c *Controller) HandleConfigUpdate(uc sun.Config) {
	log.Printf("UpdateConfig")
	c.activeConfig.Location.Lat = uc.Location.Lat
	c.activeConfig.Location.Long = uc.Location.Long
}

func (c *Controller) HandleTargetAdjustment(m sun.Message) {
	mtr := sun.MoveTargetRelative{}
	err := json.Unmarshal(m.D, &mtr)
	if err != nil {
		log.Printf("Error unmarshalling: %v", err)
	}
	//limit max movement in a single step
	maxMoveAmount := degToRad(20.0)
	if mtr.Amount > maxMoveAmount {
		mtr.Amount = maxMoveAmount
	}
	switch mtr.Direction {
	case "up":
		c.activeConfig.Target.Altitude += mtr.Amount
		if c.activeConfig.Target.Altitude > (math.Pi / 2.0) {
			c.activeConfig.Target.Altitude = math.Pi / 2.0 //cap at vertical
		}
	case "down":
		c.activeConfig.Target.Altitude -= mtr.Amount
		if c.activeConfig.Target.Altitude < 0 {
			c.activeConfig.Target.Altitude = 0 //cap at horizon
		}
	case "left":
		c.activeConfig.Target.Azimuth -= mtr.Amount
		if c.activeConfig.Target.Azimuth < -math.Pi {
			c.activeConfig.Target.Azimuth += (math.Pi * 2) //wrap around the circle
		}
	case "right":
		c.activeConfig.Target.Azimuth += mtr.Amount
		if c.activeConfig.Target.Azimuth > math.Pi {
			c.activeConfig.Target.Azimuth -= (math.Pi * 2) //wrap around the circle
		}
	default:
		c.publish <- sun.NewAckMessage(false)
	}
	log.Printf("New Target is (azi, alt) %.3f, %.3f", radToDeg(c.activeConfig.Target.Azimuth), radToDeg(c.activeConfig.Target.Altitude))
	//if we got one of the expected directions send an ack
	if mtr.Direction == "up" || mtr.Direction == "down" || mtr.Direction == "left" || mtr.Direction == "right" {
		c.publish <- sun.NewAckMessage(true)
	}
}
