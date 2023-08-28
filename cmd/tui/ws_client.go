package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	sun "github.com/mykldog7/heliostat2/pkg/types"
)

func StartConnection(address string, unlockUI *sync.Mutex) error {
	//Attempt to connect to the given address, if error pass up to caller
	conn, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		return err
	}
	//clean closure of the ws connection

	unlockUI.Unlock() //unlock the UI, allowing the UI goroutine to display the ui

	//Channel to wait for either goroutine to return an error
	errC := make(chan error)

	//Handle inward messages
	go func() {
		for {
			typ, d, err := conn.ReadMessage()
			if err != nil {
				errC <- err
			}
			if typ != websocket.TextMessage {
				errC <- fmt.Errorf("got unexpected message type %v", typ)
			}
			msg := &sun.Message{}
			err = json.Unmarshal(d, msg)
			if err != nil {
				errC <- err
			}
			switch msg.T {
			case "ActiveConfig":
				err = json.Unmarshal(d, config) //unmarshal into global config, yikes!
				if err != nil {
					errC <- err
				}
				//log.Printf("got activeConfig: %v", string(d))
			case "Ack":
				//log.Printf("got ack: %v", string(d))
			default:
				log.Printf("got unknown message type: %v with data: %v", msg.T, string(d))
			}
		}
	}()

	//Handle outward messages
	go func() {
		for {
			m := <-toServer
			if m.T == "close" {
				//close the connection gracefully
				conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "disconnect"), time.Now().Add(500*time.Millisecond))
				errC <- nil
				return //means we never write anything else
			}
			if err != nil {
				errC <- fmt.Errorf("Trouble marshalling json of %v", m)
			}
			payload, err := json.Marshal(m)
			if err != nil {
				errC <- err
			}
			//notes.SetText(string(payload))
			err = conn.WriteMessage(websocket.TextMessage, payload)
			if err != nil {
				errC <- err
			}
		}
	}()

	return <-errC //pass-up if either goroutine to return an error
}
