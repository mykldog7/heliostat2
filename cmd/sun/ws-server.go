package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type wsHandler struct {
	// logf controls where logs are sent.
	logf func(f string, v ...interface{})

	inward  chan<- interface{}
	outward chan []byte
}

func startWebsocketServer(address string, ctx context.Context, inward chan<- interface{}, outward chan []byte) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())
	s := &http.Server{
		Handler: wsHandler{
			logf:    log.Printf,
			inward:  inward,
			outward: outward,
		},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errC := make(chan error, 1)
	go func() {
		errC <- s.Serve(l)
	}()

	select {
	case err := <-errC:
		log.Printf("failed to serve: %v", err)
	case <-ctx.Done():
		return s.Shutdown(ctx)
	}
	return nil
}

func (s wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{""},
		InsecureSkipVerify: true,
	})
	if err != nil {
		s.logf("%v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "connection closing")

	log.Printf("success websocket connection, passing to handler...")

	err = s.manageWebsocketConnection(r.Context(), c)
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
		return
	}
	if err != nil {
		s.logf("failed to responde to %v: %v", r.RemoteAddr, err)
		return
	}
}

func (s wsHandler) manageWebsocketConnection(ctx context.Context, c *websocket.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//errC channel, any error terminates the connection
	errC := make(chan error)
	//reader, whenever something can be read, read it then send onwards(to control loop)
	go func() {
		for {
			typ, r, err := c.Reader(ctx)
			if err != nil {
				errC <- err
			}
			if typ != websocket.MessageText {
				s.outward <- []byte("only text(json) websocket messages can be handled")
				continue
			}

			b, err := io.ReadAll(r)
			if err != nil {
				errC <- err
			}

			msg := Message{}
			err = json.Unmarshal(b, &msg)
			if err != nil {
				s.outward <- []byte("invalid json. connection terminated.")
				c.Close(websocket.StatusUnsupportedData, "invalid json")
				errC <- err
				continue
			}

			switch msg.T {
			case "UpdateConfig":
				uc := UpdateConfig{}
				json.Unmarshal(b, &uc)
				s.inward <- uc
			default:
				log.Printf("unhandled message type:%v", msg.T)
				s.outward <- []byte(fmt.Sprintf("cant process type: %v", msg.T))
			}
		}
	}()

	//sender, anything on 'out' queue is written to websocket
	go func() {
		for {
			m := <-s.outward
			w, err := c.Writer(ctx, websocket.MessageText)
			if err != nil {
				errC <- err
			}
			w.Write(m)
			err = w.Close()
			if err != nil {
				errC <- err
			}
		}
	}()

	return <-errC
}
