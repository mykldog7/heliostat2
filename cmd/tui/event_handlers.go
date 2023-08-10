package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
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
