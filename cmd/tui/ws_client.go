package main

import (
	"context"
	"encoding/json"
	"log"

	sun "github.com/mykldog7/heliostat2/pkg/types"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WSClient struct {
	address string
	conn    *websocket.Conn
	ctx     context.Context
}

func NewConnection(ctx context.Context, address string) (WSClient, error) {
	c, _, err := websocket.Dial(ctx, address, nil)
	if err != nil {
		return WSClient{}, err
	}
	connected = true
	return WSClient{address: address, conn: c, ctx: ctx}, nil
}

// requests the current config from the controller/server
func (w *WSClient) getConfig() (*sun.Config, error) {
	payload := make(map[string]string)
	payload["t"] = "GetConfig"
	err := wsjson.Write(w.ctx, w.conn, payload)
	if err != nil {
		return nil, err
	}
	config := &sun.Config{}
	err = wsjson.Read(w.ctx, w.conn, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// updateTarget pushes new azi/alt to the server
func (w *WSClient) updateTarget(dir string, amount float64) error {
	payload := &sun.MoveTargetRelative{Direction: dir, Amount: 0.01}
	payload_bytes, _ := json.Marshal(payload)
	notes.SetText(string(payload_bytes))
	err := wsjson.Write(w.ctx, w.conn, payload)
	if err != nil {
		log.Printf("got atn err: %v", err)
		//	return err
	}
	return nil
}
