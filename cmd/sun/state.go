package main

import "time"

type State struct {
	Config   *Config   `json:"config"`
	Time     time.Time `json:"time"`
	Position struct {
		Elevation float64 `json:"ele"`
		Azimuth   float64 `json:"azi"`
	} `json:"pos"`
	Connected bool          `json:"connected"`
	Enabled   bool          `json:"enabled"`
	Uptime    time.Duration `json:"uptime"`
}
