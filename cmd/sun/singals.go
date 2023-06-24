package main

//This file contains all the types of messages that are expected to be sent and recieved from the websocket server

//Message is the wrapper for all messages, it contains a string identifying the type of message
type Message struct {
	T string                 `json:"type"`
	D map[string]interface{} `json:"-"`
}

//incoming signals (to be processed by the server)
type UpdateConfig struct {
}

//outgoing signals (to be sent to the client, status updates, etc)
