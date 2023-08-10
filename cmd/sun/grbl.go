package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

//Useful docs on details of communicating with Grbl: https://github.com/gnea/grbl/issues/822

type GrblArduino struct {
	port     serial.Port
	portName string
	mutex    sync.Mutex
}

func NewGrblArduino(ctx context.Context) (*GrblArduino, error) {
	mode := &serial.Mode{
		BaudRate: 115200, //adjust baud here, or other serial connection settings
	}

	grbl := GrblArduino{}

	//Connect to each port scanning for the one that is the grbl arduino
	err := grbl.Connect(mode)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to Grbl on %v\n", grbl.portName)

	//Control loop, manages grbl status pings.
	go func() {
		log.Printf("Grbl control loop started...")
		grblPingFreq, _ := time.ParseDuration(("1000ms"))
		statusPing := time.NewTicker(grblPingFreq)
		for {
			select {
			case <-ctx.Done():
				grbl.port.Close()
				log.Printf("Grbl control loop terminated.")
				return

			case <-statusPing.C:
				grbl.GetStatus()
				//stat, _ := grbl.GetStatus()
				//log.Printf("stat:%v", string(stat))
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
			continue
		}

		g.port = p
		g.portName = port
		g.port.SetReadTimeout(time.Second * 2)
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
	//only 1 reader/writer at a time or we get response messages confused
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if len(c) == 0 {
		log.Print("need a command to send to grbl, got nothing")
		return []byte(""), nil
	}
	// write command
	n, err := g.port.Write(c)
	if n != len(c) || err != nil {
		return []byte(""), fmt.Errorf("error writing to Grbl, or unexpected number of bytes")
	}
	// read response
	resp, err := g.readLine()
	if err != nil {
		return nil, err
	}
	//did the response contain ok?
	if bytes.Equal(resp, []byte("ok\r\n")) {
		return resp, nil
	}
	return resp, fmt.Errorf("grbl reports error: %v", string(resp))
}

// GrblReadBanner is used to read the startup banner.
func (g *GrblArduino) GrblReadBanner() error {
	//only 1 reader/writer at a time or we get response messages confused
	g.mutex.Lock()
	defer g.mutex.Unlock()
	//storage to gather bytes
	_, err := g.port.Write([]byte{0xd, 0xa, 0xd, 0xa}) // wake up grbl
	time.Sleep(2 * time.Second)                        // give it a chance to come up
	if err != nil {
		return err
	}
	var line []byte
	line, err = g.readLine()
	if err != nil {
		return err
	}
	if strings.Contains(string(line), "Grbl") {
		return nil
	}
	return fmt.Errorf("no Grbl banner found")
}

// readLine reads a line(terminated by /r/n) from grbl, used when a real-time command causes grbl to print the current status
func (g *GrblArduino) readLine() ([]byte, error) {
	//read a byte and append it if available.. when a \r\n is recieved we have the banner
	buff := make([]byte, 1)
	line := make([]byte, 0, 80)
	for {
		n, err := g.port.Read(buff)
		//log.Printf("new bytes:%v, line:%v", n, string(line))
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, fmt.Errorf("no bytes from Grbl before timeout")
		}
		if n > 0 {
			line = append(line, buff...)
			//have we collected more than 2 bytes?
			if len(line) >= 2 {
				//if the line is "\r\n" then zero it and read the next line instead
				if len(line) == 2 && bytes.Equal(line, []byte("\r\n")) {
					line = make([]byte, 0, 80)
					continue
				}
				//are the 'last' two bytes an end of line marker?
				if bytes.Equal(line[len(line)-2:], []byte("\r\n")) {
					return line, nil
				}
			}
		}
	}
}

// GetStatus sends the real-time command '?' to grbl and reports the response.
func (g *GrblArduino) GetStatus() ([]byte, error) {
	//only 1 reader/writer at a time or we get response messages confused
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.port == nil {
		return []byte{}, fmt.Errorf("can't get status if not connected")
	}
	g.port.Write([]byte("?"))
	line, err := g.readLine()
	if err != nil {
		return []byte{}, err
	}
	return line, nil
}
