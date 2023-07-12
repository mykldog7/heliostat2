package main

import (
	"log"

	"github.com/256dpi/gcode"
)

// PositionToGCode builds a GCode command to send the mirror to the desired azi/ele
func PositionToGCode(azi float64, ele float64) []byte {
	if azi > 360.0 || azi < 0.0 {
		log.Printf("Azimuth out of range can't convert to GCode: %v", azi)
		return []byte{}
	}
	if ele > 90.0 || ele < 0.0 {
		log.Printf("Elevation out of range can't convert to GCode: %v", ele)
		return []byte{}
	}
	line := gcode.Line{
		Codes: make([]gcode.GCode, 0, 2),
	}
	line.Codes = append(line.Codes, gcode.GCode{Letter: "X", Value: azi})
	line.Codes = append(line.Codes, gcode.GCode{Letter: "Y", Value: ele})
	return []byte(line.String())
}
