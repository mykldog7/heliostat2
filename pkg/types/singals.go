package types

import "time"

//This file contains all the types of messages that are expected to be sent and recieved from the websocket server

//Message is the wrapper for all messages, it contains a string identifying the type of message and the message data
type Message struct {
	T string `json:"t"`
	D []byte
}

//incoming signals (to be processed by the server)
//Config <- see config.go
//{"t":"config", "loc":{"lat":1,"long":100}}

//Used to move the target in a particular direction
type MoveTargetRelative struct {
	Direction string  `json:"direction"` //expected up down left, right
	Amount    float64 `json:"radians"`
}

//Provide an immediate 'override' time to the system(it should have its own internal clock, but this can override that)
type SetTime struct {
	Time time.Time `json:"datetime"`
}

//Allow a previously set 'override' time to be removed, reverting to 'system'/'background' time.
type ResetTime struct {
}

//Request for a TargetPosition response
type GetTargetPosition struct {
}

type SetUpdateFreq struct {
	Period time.Duration `json:"period"`
}

//outgoing signals (to be sent to the client, status updates, etc)
type Ack struct {
	Success bool `json:"success"`
}

func NewAck(s bool) Ack { return Ack{Success: s} }

//Outward signals, published by the robot to ws subscribers

//Used to give the current target position/coordinates
type TargetPosition struct {
	Azimuth  float64 `json:"azi"`
	Altitude float64 `json:"alt"`
}

//sent whenever the mirror repositions
type Reposition struct {
	Time      time.Time `json:"time"`
	Azimuth   float64   `json:"azi"`
	Elevation float64   `json:"ele"`
}

type Status struct {
	Message string `json:"msg"`
}

func NewStatus(s string) Status { return Status{Message: s} }
