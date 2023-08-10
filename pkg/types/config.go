package types

import (
	"encoding/json"
	"time"
)

// Config is used to store the core configuration of the heliostat at the present time.
type Config struct {
	TimeProgression float64   `json:"progression_factor"`
	OverrideTime    time.Time `json:"override_time"` // the 'sim' or override time
	Location        Location  `json:"loc"`
	Target          struct {
		Altitude float64 `json:"alt"`
		Azimuth  float64 `json:"azi"`
	} `json:"target"`
}

// Location stores a particular point on the earths surface
type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func (c Config) String() string {
	s, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(s)
}
