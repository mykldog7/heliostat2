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
	bytes, err := json.Marshal(c.activeConfig)
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
	}
}
