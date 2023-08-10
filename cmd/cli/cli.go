package main

import (
	"log"
	"math"
	"time"

	"github.com/sixdouglas/suncalc"
)

func main() {
	//theTime := time.Date(2023, 7, 1, 14, 30, 45, 100, time.Local)
	theTime := time.Date(2023, 7, 18, 20, 55, 0, 0, time.Local)
	timeSun(theTime)
	theNextTime := theTime.Add(time.Hour * 12)
	timeSun(theNextTime)
}
func timeSun(t time.Time) {
	log.Println("The time is", t)
	sL := suncalc.GetPosition(t, -36.9565, 174.7777)
	log.Printf("Sun is at: %v, %v", radToDeg(sL.Azimuth), radToDeg(sL.Altitude))
}

// Utility functions
func radToDeg(a float64) float64 { return a * 180.0 / math.Pi }
func degToRad(a float64) float64 { return a * math.Pi / 180.0 }
