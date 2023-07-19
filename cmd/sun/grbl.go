package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

//Useful docs on details of communicating with Grbl: https://github.com/gnea/grbl/issues/822

type GrblArduino struct {
	fromGrbl chan []byte
	toGrbl   chan []byte
	port     serial.Port
	portName string
}

func NewGrblArduino(ctx context.Context) (*GrblArduino, error) {
	mode := &serial.Mode{
		BaudRate: 115200, //adjust baud here, or other serial connection settings
	}

	grbl := GrblArduino{
		toGrbl:   make(chan []byte),
		fromGrbl: make(chan []byte),
	}

	//Connect to each port scanning for the one that is the grbl arduino
	err := grbl.Connect(mode)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to Grbl on %v\n", grbl.portName)

	//Control loop, writes to fromGrbl, sends from toGrbl, terminates on ctx.Done
	go func() {
		log.Printf("Grbl control loop started...")
		for {
			select {
			case <-ctx.Done():
				grbl.port.Close()
				log.Printf("Grbl control loop terminated.")
				return
			case msg := <-grbl.toGrbl:
				resp, err := grbl.GrblSendCommandGetResponse(msg)
				if err != nil {
					log.Fatal(err)
				}
				if !bytes.Equal(resp, []byte("ok")) {
					log.Printf("!ok response from grbl: %v", resp)
				}
				grbl.fromGrbl <- resp
			}
		}
	}()
	return &grbl, nil
}

// Connect connects to all availabl serial ports and listens to each to see if it produces the grbl banner
// when it does, that serial connection is left open, the Grbl is now connected and ready
func (g *GrblArduino) Connect(mode *serial.Mode) error {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
		return err
	}
	if len(ports) == 0 {
		return fmt.Errorf("no serial ports available, is grbl(arduino) connnected?")
	}
	for _, port := range ports {
		log.Printf("Checking port: %v to see if it is \"Grbl\"...", port)
		p, err := serial.Open(port, mode)
		if err != nil {
			log.Printf("error with port %v: %v", port, err)
			log.Printf("trying next serial port")
			continue
		}
		g.port = p
		g.portName = port
		err = g.GrblReadBanner()
		if err != nil {
			log.Printf("error detecting banner on port %v: %v", port, err)
			continue
		}
		return nil //successful exit, g is now connected
	}
	return fmt.Errorf("could not find a serial port with Grbl's banner")
}

func (g *GrblArduino) GrblSendCommandGetResponse(c []byte) ([]byte, error) {
	if len(c) == 0 {
		log.Print("need a command to send to grbl, got nothing")
		return []byte(""), nil
	}
	n, err := g.port.Write(c)
	if n != len(c) || err != nil {
		return []byte(""), fmt.Errorf("error writing to Grbl, or unexpected number of bytes")
	}
	//storage to gather response
	buff := make([]byte, 1)
	line := make([]byte, 0, 82)
	lineStatus := make([]byte, 0, 4)
	lineComplete := false
	//ready a byte and append it if available.. when a \r\n is recieved we have a 'line' then check if its 'ok|error'
	for {
		n, _ := g.port.Read(buff)
		if n == 0 {
			continue
		}
		switch lineComplete {
		case false:
			line = append(line, buff...)
			if len(line) > 2 {
				if bytes.Equal(line[len(line)-2:], []byte("\r\n")) {
					lineComplete = true
				}
			}
		case true:
			lineStatus = append(lineStatus, buff...)
			if len(lineStatus) > 2 {
				if bytes.Equal(lineStatus[len(lineStatus)-2:], []byte("\r\n")) {
					if bytes.Equal(lineStatus, []byte("ok\r\n")) {
						return line, nil
					} else {
						return line, fmt.Errorf("grbl reports error: %v", string(lineStatus))
					}
				}
			}
		}
	}
}

// GrblReadBanner is used to read the startup banner.
func (g *GrblArduino) GrblReadBanner() error {
	//storage to gather bytes
	buff := make([]byte, 1)
	line := make([]byte, 0, 82)
	_, err := g.port.Write([]byte("\r\n\r\n")) // wake up grbl
	time.Sleep(2 * time.Second)                // give it a chance to come up
	if err != nil {
		return err
	}
	//read a byte and append it if available.. when a \r\n is recieved we have the banner
	for {
		g.port.SetReadTimeout(time.Second * 2)
		n, err := g.port.Read(buff)
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("no bytes from Grbl before timeout")
		}
		if n > 0 {
			line = append(line, buff...)
			//have we collected more than 2 bytes?
			if len(line) > 2 {
				//are the 'last' two bytes an end of line marker?
				if bytes.Equal(line[len(line)-2:], []byte("\r\n")) {
					//does the line contain the 'Grbl' banner?
					if strings.Contains(string(line), "Grbl") {
						return nil
					}
					return fmt.Errorf("no Grbl banner found")
				}
			}
		}
	}
}
