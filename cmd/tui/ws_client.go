package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	sun "github.com/mykldog7/heliostat2/pkg/types"
	"nhooyr.io/websocket"
)

func StartConnection(ctx context.Context, address string) error {
	conn, _, err := websocket.Dial(ctx, address, nil) // set the global connection
	if err != nil {
		return err
	}
	connected = true
	//Handle incoming messages
	for {
		typ, r, err := conn.Reader(ctx)
		if err != nil {
			return err
		}
		if typ != websocket.MessageText {
			return fmt.Errorf("got unexpected message type %v", typ)
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		msg := &sun.Message{}
		err = json.Unmarshal(data, msg)
		if err != nil {
			return err
		}
		switch msg.T {
		case "ActiveConfig":
			err = json.Unmarshal(data, config) //unmarshal into global config, yikes!
			if err != nil {
				return err
			}
		case "Ack":
			log.Printf("got ack: %v", string(data))
		default:
			log.Printf("got unknown message type: %v", msg.T)
		}
	}
}

// updateTarget pushes new azi/alt to the server
func updateTarget(dir string, amount float64) error {
	payload := &sun.MoveTargetRelative{Direction: dir, Amount: 0.01}
	payload_bytes, _ := json.Marshal(payload)
	//notes.SetText(string(payload_bytes))
	log.Printf("sending: %v", string(payload_bytes))
	conn.Write(ctx, websocket.MessageText, payload_bytes)
	return nil
}
