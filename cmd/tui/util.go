package main

import "math"

// Utility functions
func radToDeg(a float64) float64 { return a * 180.0 / math.Pi }
func degToRad(a float64) float64 { return a * math.Pi / 180.0 }
