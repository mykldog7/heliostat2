package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	sun "github.com/mykldog7/heliostat2/pkg/types"
	"nhooyr.io/websocket"
)

func StartConnection(ctx context.Context, address string, unlockUI *sync.Mutex) error {
	conn, _, err := websocket.Dial(ctx, address, nil) // set the global connection
	if err != nil {
		return err
	}
	unlockUI.Unlock() //unlock the UI, allowing the UI goroutine to proceed

	//Channel to wait for either goroutine to return an error
	errC := make(chan error)

	//Handle incoming messages
	go func() {
		for {
			typ, r, err := conn.Reader(ctx)
			if err != nil {
				errC <- err
			}
			if typ != websocket.MessageText {
				errC <- fmt.Errorf("got unexpected message type %v", typ)
			}
			data, err := io.ReadAll(r)
			if err != nil {
				errC <- err
			}
			msg := &sun.Message{}
			err = json.Unmarshal(data, msg)
			if err != nil {
				errC <- err
			}
			switch msg.T {
			case "ActiveConfig":
				err = json.Unmarshal(data, config) //unmarshal into global config, yikes!
				if err != nil {
					errC <- err
				}
				log.Printf("got activeConfig: %v", string(data))
			case "Ack":
				log.Printf("got ack: %v", string(data))
			default:
				log.Printf("got unknown message type: %v", msg.T)
			}
		}
	}()

	//Handle outgoing messages
	go func(ctx context.Context) {
		for {
			m := <-toServer
			log.Printf("Sending message: %v", string(m))
			w, err := conn.Writer(ctx, websocket.MessageText)
			if err != nil {
				errC <- err
			}
			w.Write(m)
			err = w.Close()
			if err != nil {
				errC <- err
			}
		}
	}(ctx)

	return <-errC //wait for either goroutine to return an error
}
