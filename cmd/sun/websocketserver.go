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

// websocket handler is instantiated each time we have a client connect
// it must register its outward channel with the SubManager to get outward messages
type wsHandler struct {
	// logf controls where logs are sent.
	logf    func(f string, v ...interface{})
	inward  chan<- interface{}
	manager *SubManager
}

func startWebsocketServer(address string, ctx context.Context, inward chan<- interface{}, publish chan []byte) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())
	sm := NewSubManger(publish)
	sm.Start()
	s := &http.Server{
		Handler: wsHandler{
			logf:    log.Printf,
			inward:  inward,
			manager: sm,
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
		s.logf("closed websocket")
		return
	}
	if err != nil {
		s.logf("error with websocket to %v: %v", r.RemoteAddr, err)
		return
	}
}

func (s wsHandler) manageWebsocketConnection(ctx context.Context, c *websocket.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//create outward channel, and register it to receive 'publish' messages
	outward := make(chan []byte)
	s.manager.AddSub(outward)

	//start a waiting go routine, when the ctx is cancelled(ie we have the end of the connection, then unregister our channel)
	go func() {
		<-ctx.Done()
		s.manager.RemoveSub(outward)
	}()

	//errC channel, any error terminates the connection
	errC := make(chan error)

	//sender, anything on 'out' queue is written to websocket
	go func() {
		for {
			m := <-outward
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

	//reader, whenever something can be read, read it then send onwards(to control loop)
	go func() {
		for {
			typ, r, err := c.Reader(ctx)
			if err != nil {
				errC <- err
			}
			if typ != websocket.MessageText {
				outward <- []byte("only text(json) websocket messages can be handled")
				continue
			}

			b, err := io.ReadAll(r)
			if err != nil {
				errC <- err
			}

			msg := Message{}
			err = json.Unmarshal(b, &msg)
			if err != nil {
				outward <- []byte("unhandled: need valid json")
				continue
			}

			switch msg.T {
			case "config":
				uc := Config{}
				json.Unmarshal(b, &uc)
				s.inward <- uc
			default:
				log.Printf("unhandled message type:%v", msg.T)
				outward <- []byte(fmt.Sprintf("cant process type(check \"t\" key?): %v", msg.T))
			}
		}
	}()

	return <-errC
}
