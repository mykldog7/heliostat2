package main

import (
	"log"
	"math"
)

// HandleConfigUpdate updates the 'activeconfig on the controller
func (c *Controller) HandleConfigUpdate(uc Config) {
	log.Printf("UpdateConfig")
	c.activeConfig.Location.Lat = uc.Location.Lat
	c.activeConfig.Location.Long = uc.Location.Long
}

func (c *Controller) HandleTargetAdjustment(m MoveTargetRelative) {
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
