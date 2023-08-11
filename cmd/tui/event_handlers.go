package main

import (
	"encoding/json"
	"log"

	"github.com/gdamore/tcell/v2"
	sun "github.com/mykldog7/heliostat2/pkg/types"
)

func adjustTargetEventHandler(e *tcell.EventKey) *tcell.EventKey {
	key, ch := e.Key(), e.Rune()
	if key == tcell.KeyRune {
		//log.Printf("rune:%v", ch)
		switch ch {
		case 'w':
			err := updateTarget("up", currentMoveSize)
			if err != nil {
				log.Fatalf("updateT err: %v", err)
			}
		case 'a':
			err := updateTarget("left", currentMoveSize)
			if err != nil {
				log.Fatalf("updateT err: %v", err)
			}
		case 's':
			err := updateTarget("down", currentMoveSize)
			if err != nil {
				log.Fatalf("updateT err: %v", err)
			}
		case 'd':
			err := updateTarget("right", currentMoveSize)
			if err != nil {
				log.Fatalf("updateT err: %v", err)
			}
		case '<':
			currentMoveSize *= 10.0
		case '>':
			currentMoveSize *= 0.1
		}
	}
	return e
}

// updateTarget pushes new azi/alt to the server
func updateTarget(dir string, amount float64) error {
	payload := &sun.MoveTargetRelative{Direction: dir, Amount: 0.01}
	payload_bytes, _ := json.Marshal(payload)
	//notes.SetText(string(payload_bytes))
	toServer <- payload_bytes
	return nil
}
