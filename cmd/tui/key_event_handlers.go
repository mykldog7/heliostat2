package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	sun "github.com/mykldog7/heliostat2/pkg/types"
)

func adjustTargetEventHandler(e *tcell.EventKey) *tcell.EventKey {
	key, ch := e.Key(), e.Rune()
	if key == tcell.KeyRune {
		switch ch {
		case 'w':
			err := updateTarget("up", currentMoveSize)
			if err != nil {
				log.Fatalf("err: %v", err)
			}
			return nil
		case 'a':
			err := updateTarget("left", currentMoveSize)
			if err != nil {
				log.Fatalf("err: %v", err)
			}
			return nil
		case 's':
			err := updateTarget("down", currentMoveSize)
			if err != nil {
				log.Fatalf("err: %v", err)
			}
			return nil
		case 'd':
			err := updateTarget("right", currentMoveSize)
			if err != nil {
				log.Fatalf("err: %v", err)
			}
			return nil
		case '<':
			currentMoveSize *= 2.0
			notes.SetText(fmt.Sprintf("Move size(degrees): %v", radToDeg(currentMoveSize)))
			return nil
		case '>':
			currentMoveSize *= 0.5
			notes.SetText(fmt.Sprintf("Move size(degrees): %v", radToDeg(currentMoveSize)))
			return nil
		default:
			return e
		}
	}
	return e
}

// updateTarget pushes new azi/alt to the server
func updateTarget(dir string, amount float64) error {
	payload := sun.MoveTargetRelative{Direction: dir, Amount: amount}
	payload_bytes, _ := json.Marshal(payload)
	toServer <- sun.Message{T: "MoveTargetRelative", D: payload_bytes}
	return nil
}
