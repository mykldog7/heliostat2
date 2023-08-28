package main

import (
	"fmt"
	"math"

	"github.com/256dpi/gcode"
)

// PositionToGCode builds a GCode command to send the mirror to the desired azi/ele
// azimuth -pi/2 =>  x=0
// azimuth increases, x increases
func PositionToGCode(azi float64, ele float64) ([]byte, error) {
	//check valid azimuth
	if azi > 180.0 || azi < -180.0 {
		return nil, fmt.Errorf("unexpected value for azimuth: %v", azi)
	}
	if ele > 90.0 || ele < 0.0 {
		return nil, fmt.Errorf("unexpected value for elevation: %v", ele)
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
//read up here: https://mathworld.wolfram.com/SphericalCoordinates.html

func calculateMirrorTarget(pAzi float64, pAlt float64, tAzi float64, tAlt float64) (float64, float64) {
	//convert to cartesian vectors
	px, py, pz := toCartesianCoords(pAzi, pAlt, 1.0)
	tx, ty, tz := toCartesianCoords(tAzi, tAlt, 1.0)

	//add the vectors
	mx := px + tx
	my := py + ty
	mz := pz + tz

	//convert back to polar/spherical co-oridnates
	mAzi, mAlt, _ := toSphericalCoords(mx, my, mz)

	return mAzi, mAlt
}

// toCartesianCoords returns the cartesian coordinates of the given spherical coordinates, note: this flips the zenith angle
func toCartesianCoords(azi float64, alt float64, r float64) (float64, float64, float64) {
	//alt is in latitude, convert to polar angle: zenith is zero not 90.
	phi := (math.Pi / 2) - alt

	x := r * math.Cos(azi) * math.Sin(phi)
	y := r * math.Sin(azi) * math.Sin(phi)
	z := r * math.Cos(phi)
	return x, y, z
}

// toSphericalCoords returns the sperical coordinates of the given carteisan coords
func toSphericalCoords(x float64, y float64, z float64) (float64, float64, float64) {
	r := math.Sqrt((math.Pow(x, 2.0) + math.Pow(y, 2.0) + math.Pow(z, 2.0)))
	azi := math.Atan2(y, x)
	phi := math.Acos(z / r)

	//alt is in latitude, convert from polar angle: zenith is 90, but the formula assumes 90
	alt := (math.Pi / 2) - phi
	return azi, alt, r
}
