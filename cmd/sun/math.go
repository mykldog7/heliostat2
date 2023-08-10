package main

import (
	"fmt"
	"math"

	"github.com/256dpi/gcode"
)

// PositionToGCode builds a GCode command to send the mirror to the desired azi/ele
func PositionToGCode(azi float64, ele float64) ([]byte, error) {
	if azi > 180.0 || azi < -180.0 {
		return nil, fmt.Errorf("unexpected value for azimuth: %v", azi)
	}
	if ele > 90.0 || ele < 0.0 {
		return nil, fmt.Errorf("unexpected value for elevation: %v", ele)
	}
	//adjust Azimuth to range 0 -> 360 instead of -180 -> 180... this will allow us to 'zero' the mirror against a 'hard' stop
	if azi < 0 {
		azi = 360 + azi
	}

	line := gcode.Line{
		Codes: make([]gcode.GCode, 0, 2),
	}
	line.Codes = append(line.Codes, gcode.GCode{Letter: "X", Value: azi})
	line.Codes = append(line.Codes, gcode.GCode{Letter: "Y", Value: ele})
	return []byte(line.String()), nil
}

// Utility functions
func radToDeg(a float64) float64 { return a * 180.0 / math.Pi }
func degToRad(a float64) float64 { return a * math.Pi / 180.0 }

// calculateMirrorTarget returns the midpoint at which to point the mirror's normal
// This is achieved by converting the polar coordinates to cartesian, and then adding the unit vectors,
// and finally converting back to polar cooridnates
func calculateMirrorTarget(pAzi float64, pAlt float64, tAzi float64, tAlt float64) (float64, float64) {
	//convert to cartesian vectors
	pIAlt := (math.Pi / 2) - pAlt
	pX := math.Sin(pIAlt) * math.Cos(pAzi)
	pY := math.Sin(pIAlt) * math.Sin(pAzi)
	pZ := math.Cos(pIAlt)

	tIAlt := (math.Pi / 2) - tAlt
	tX := math.Sin(tIAlt) * math.Cos(tAzi)
	tY := math.Sin(tIAlt) * math.Sin(tAzi)
	tZ := math.Cos(tIAlt)

	//add the vectors
	mX := pX + tX
	mY := pY + tY
	mZ := pZ + tZ

	//log.Printf("vector (%5.4f,%5.4f,%5.4f) vector (%5.4f,%5.4f,%5.4f) vector (%5.4f,%5.4f,%5.4f)\n", pX, pY, pZ, tX, tY, tZ, mX, mY, mZ)

	//convert back to polar/spherical co-oridnates
	mAzi := math.Atan(mY / mX)
	mAlt := math.Atan(math.Sqrt(math.Pow(mX, 2.0)+math.Pow(mY, 2.0)) / mZ)
	mIAlt := 0.0
	if mAlt >= 0 && mAlt <= (math.Pi/2.0) {
		mIAlt = (math.Pi / 2.0) - mAlt
	}
	return mAzi, mIAlt
}
