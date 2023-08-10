package main

import (
	"time"

	"github.com/mykldog7/heliostat2/pkg/types"
)

type State struct {
	Config   *types.Config `json:"config"`
	Time     time.Time     `json:"time"`
	Position struct {
		Elevation float64 `json:"ele"`
		Azimuth   float64 `json:"azi"`
	} `json:"pos"`
	Connected bool          `json:"connected"`
	Enabled   bool          `json:"enabled"`
	Uptime    time.Duration `json:"uptime"`
}
