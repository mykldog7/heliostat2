package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/mykldog7/heliostat2/pkg/types"

	"github.com/sixdouglas/suncalc"
)

type Controller struct {
	activeConfig      types.Config
	in                <-chan types.Message
	publish           chan<- []byte
	updatePeriod      time.Duration
	localTime         time.Time // the 'real' time
	lastUpdate        time.Time // When was the lastUpdate completed
	usingOverrideTime bool      // which time are we using for calculations
	grbl              *GrblArduino
}

func NewController(inChan <-chan types.Message, outChan chan<- []byte, grbl *GrblArduino) Controller {
	defaultPeriod, _ := time.ParseDuration("10s")
	defaultLat, defaultLong := 174.8860, -40.9006
	initialTime := time.Date(2023, 1, 1, 12, 00, 0, 0, time.Local)
	return Controller{
		activeConfig: types.Config{
			Location:        types.Location{Lat: defaultLat, Long: defaultLong},
			OverrideTime:    initialTime,
			TimeProgression: 60.0,
			Target: struct {
				Altitude float64 "json:\"alt\""
				Azimuth  float64 "json:\"azi\""
			}{Altitude: 10.0, Azimuth: 20.0}},
		in:                inChan,
		publish:           outChan,
		updatePeriod:      defaultPeriod,
		localTime:         time.Now(),
		lastUpdate:        time.Now(),
		usingOverrideTime: true, //TODO back to false
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
			//process incoming control messages
			log.Printf("Recieved inwards message with type: %v", msg.T)

			switch msg.T {

			case "UpdateConfig":
				uc := types.Config{}
				json.Unmarshal(msg.D, &uc)
				c.HandleConfigUpdate(uc)

			case "GetConfig":
				c.HandleGetActiveConfig()

			case "move":
				mtr := types.MoveTargetRelative{}
				json.Unmarshal(msg.D, &mtr)
				c.HandleTargetAdjustment(mtr)

			default:
				log.Printf("Controller dropped message with type %v as no handler defined.", msg.T)
			}

		case <-ticker.C:
			//updates controller times
			c.localTime = time.Now()
			gap := time.Since(c.lastUpdate)
			gap *= time.Duration(c.activeConfig.TimeProgression)
			c.activeConfig.OverrideTime = c.activeConfig.OverrideTime.Add(gap)
			c.lastUpdate = time.Now()

			//recalculate desired position
			mAzi, mAlt := c.RecalculateDesiredMirrorPosition()
			log.Printf("Calculated Desired Mirror Position: azi:%v, alt:%v", mAzi, mAlt)
			mAzi_Deg := radToDeg(mAzi)
			mAlt_Deg := radToDeg(mAlt)
			//convert position to GCode..
			code, err := PositionToGCode(mAzi_Deg, mAlt_Deg)
			if err != nil {
				log.Printf("%v", err)
			}
			resp, err := c.grbl.GrblSendCommandGetResponse(code)
			if err != nil {
				log.Printf("Error from GRBL: %v", err)
				return err
			}
			log.Printf("Sent %v to grbl for moment %v ... got response \"%v\"", strings.TrimSuffix(string(code), "\n"), c.cTime(), string(resp[0:2]))
			//c.publish <- []byte(fmt.Sprintf("Sent %v to grbl at: %v", string(code), tick)) //this will generally cause problems for the clients, if they are expecting something else
		}
	}
}

// RecalculateTarget returns the position correct according to configured time
// Altitude: sun altitude above the horizon in radians, e.g. -1 at the horizon and PI/2 at the zenith (straight over your head)
// Azimuth: sun azimuth in radians (direction along the horizon, measured from south to west), e.g. -1 is south and Math.PI * 3/4 is northwest
func (c *Controller) RecalculateDesiredMirrorPosition() (float64, float64) {
	sun := suncalc.GetPosition(c.cTime(), c.activeConfig.Location.Lat, c.activeConfig.Location.Long)
	mirrorAzi, mirrorAlt := calculateMirrorTarget(sun.Azimuth, sun.Altitude, c.activeConfig.Target.Azimuth, c.activeConfig.Target.Altitude)
	return mirrorAzi, mirrorAlt
}

// returns the 'current-active' time
func (c *Controller) cTime() time.Time {
	if c.usingOverrideTime {
		return c.activeConfig.OverrideTime
	}
	return c.localTime
}
