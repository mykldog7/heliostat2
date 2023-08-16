package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mykldog7/heliostat2/pkg/types"
)

// websocket handler is instantiated each time we have a client connect
// it must register its outward channel with the SubManager to get outward messages
type wsHandler struct {
	// logf controls where logs are sent.
	inward  chan<- types.Message
	manager *SubManager
}

func startWebsocketServer(address string, ctx context.Context, inward chan<- types.Message, publish chan []byte) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())
	sm := NewSubManger(publish)
	s := &http.Server{
		Handler: wsHandler{
			inward:  inward,
			manager: sm,
		},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errC := make(chan error, 1)
	//start SubManager
	go func() {
		errC <- sm.Start()
	}()
	//start HTTP server
	go func() {
		errC <- s.Serve(l)
	}()

	select {
	case err := <-errC:
		log.Printf("failed with error: %v", err)
		return err
	case <-ctx.Done():
		return s.Shutdown(ctx)
	}
}

func (s wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close()
	id, _ := uuid.NewUUID()

	log.Printf("Success websocket upgrade... connection id:%v", id)

	err = s.manageWebsocketConnection(r.Context(), c, id)
	if err != nil {
		log.Printf("error with websocket to %v: %v", r.RemoteAddr, err)
		return
	}
}

func (s wsHandler) manageWebsocketConnection(ctx context.Context, c *websocket.Conn, id uuid.UUID) error {
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
	go func(c *websocket.Conn) {
		for {
			m := <-outward
			log.Printf("Write message: %v", string(m))
			err := c.WriteMessage(websocket.TextMessage, m)
			if err != nil {
				errC <- err
			}
		}
	}(c)

	//reader, whenever something can be read, read it then drop or send onwards(to control loop)
	go func(c *websocket.Conn) {
		for {
			typ, b, err := c.ReadMessage()
			if websocket.IsCloseError(err, 1000) {
				log.Printf("Client diconnected. connection id:%v", id)
				errC <- nil
				return
			}
			if typ != websocket.TextMessage {
				outward <- []byte("{\"error\":\"only text(json) websocket messages can be handled\"}")
				continue
			}
			if err != nil {
				errC <- err
			}
			log.Printf("Read message: %v", string(b))

			msg := types.Message{}
			err = json.Unmarshal(b, &msg)
			if err != nil {
				outward <- []byte("{\"error\":\"unhandled input: need valid json with 't' key specifying a valid type\"}")
				continue //not a message we can work with, skip and continue
			}
			msg.D = b       //save data into message so it can be unmarshalled again later
			s.inward <- msg //send the msg to the controller to be handled
		}
	}(c)

	return <-errC
}
