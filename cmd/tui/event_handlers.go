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
			err := client.updateTarget("up", currentMoveSize)
			if err != nil {
				log.Fatalf("updateT err: %v", err)
			}
			//app.QueueUpdateDraw(displayAdjustTarget)
		case 'a':
			client.updateTarget("left", currentMoveSize)
		case 's':
			client.updateTarget("down", currentMoveSize)
		case 'd':
			client.updateTarget("right", currentMoveSize)
		case '<':
			currentMoveSize *= 10.0
		case '>':
			currentMoveSize *= 0.1
		}
	}
	return e
}
