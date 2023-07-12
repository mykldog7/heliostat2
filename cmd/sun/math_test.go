package main

import (
	"math"
	"testing"
)

// testCase defines the target, sun, and expected mirror position(the mid angle)
type testCase struct {
	targetAzi, targetAlt, sunAzi, sunAlt, midAzi, midAlt float64
}

var cases = []testCase{
	{0, 0, math.Pi / 4.0, 0, math.Pi / 8.0, 0},
	{0, math.Pi / 2.0, 0, 0, 0, math.Pi / 4.0},
	{math.Pi, math.Pi / 4.0, 0, math.Pi / 4.0, math.Pi / 2.0, math.Pi / 4.0},
}

func TestMath(tt *testing.T) {
	for _, t := range cases {
		if rAzi, rAlt := calculateMirrorTarget(t.targetAzi, t.targetAlt, t.sunAzi, t.sunAlt); rAzi != t.midAzi && rAlt != t.midAlt {
			tt.Errorf("Error with TestMath... Got: %v,%v expected: %v,%v", rAzi, rAlt, t.midAzi, t.midAlt)
		}
	}
}
