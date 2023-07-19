package main

import (
	"encoding/json"
)

// Config is used to store the core configuration of the heliostat at the present time.
// example: {"name":"test", "loc":{"lat": 130, "long": -42}}
type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}
type Config struct {
	Location Location `json:"loc"`
	Target   struct {
		Altitude float64 `json:"alt"`
		Azimuth  float64 `json:"azi"`
	}
}

func (c Config) String() string {
	s, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(s)
}
