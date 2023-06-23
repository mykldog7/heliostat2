package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func startWebsocketServer(address string, ctx context.Context, inward chan<- interface{}, outward <-chan interface{}) error {
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
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case <-ctx.Done():
		return s.Shutdown(ctx)
	}
}

type wsHandler struct {
	// logf controls where logs are sent.
	logf    func(f string, v ...interface{})
	inward  chan<- interface{}
	outward <-chan interface{}
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

	l := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	for {
		err = manageWebsocketConnection(r.Context(), c, l)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			s.logf("failed to echo with %v: %v", r.RemoteAddr, err)
			return
		}
	}
}

func manageWebsocketConnection(ctx context.Context, c *websocket.Conn, l *rate.Limiter) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	err := l.Wait(ctx)
	if err != nil {
		return err
	}

	incoming_config := Config{}
	fmt.Printf("%v", incoming_config)
	err = wsjson.Read(ctx, c, &incoming_config)
	if err != nil {
		return err
	}
	fmt.Printf("%v", incoming_config)

	typ, r, err := c.Reader(ctx)
	if err != nil {
		return err
	}

	w, err := c.Writer(ctx, typ)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to io.Copy: %w", err)
	}

	err = w.Close()
	return err
}
