package main

import "testing"

//testCase defines the target, sun, and expected mirror position(the mid angle)
type testCase struct {
	targetAzi, targetAlt, sunAzi, sunAlt, midAzi, midAlt float64
}

var cases = []testCase{
	testCase{0, 0, 90, 10, 45, 5},
	testCase{0, 0, 90, 10, 45, 5},
}

func TestMath(tt *testing.T) {
	for _, t := range cases {
		if rAzi, rAlt := calculateMirrorTarget(t.targetAzi, t.targetAlt, t.sunAzi, t.sunAlt); rAzi != t.midAzi && rAlt != t.midAlt {
			tt.Errorf("Error with TestMath... Got: %v,%v expected: %v,%v", rAzi, rAlt, t.midAzi, t.midAlt)
		}
	}
}
