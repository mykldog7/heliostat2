package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func runUI() error {
	app = tview.NewApplication()

	//actions available in the main menu
	actions = tview.NewList().
		AddItem("Quit", "close app", 'q', func() { app.Stop() }).
		AddItem("Adjust Target", "move the target with w,a,s,d", 'm', displayAdjustTarget).
		AddItem("Adjust Lat/Long", "set the mirror lat, long", 'l', displayLatLong).
		AddItem("Configure Time", "override the machine time", 'o', displayAdjustTime)
	actions.SetBorder(true).SetTitle("Available Actions")

	details = tview.NewFlex()
	details.SetBorder(true)
	details.AddItem(tview.NewTextArea().SetText("\"Tui\" used to manage a heliostat server/controller by sending messages over a websocket.", false), 0, 1, false)

	connectionStatus := ""
	if connected {
		connectionStatus = fmt.Sprintf("Connected to %v", *address)
	} else {
		connectionStatus = "not connected"
	}
	menuFrame := tview.NewFrame(actions).SetBorders(0, 0, 0, 0, 0, 0).AddText(connectionStatus, false, tview.AlignCenter, tcell.ColorDarkGreen)

	menuLayout := tview.NewFlex().
		AddItem(menuFrame, 0, 1, true).
		AddItem(details, 0, 1, false)

	app.SetRoot(menuLayout, true)
	//add an 'esc' key handler at the top level to return to the menu
	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			if actions.HasFocus() {
				app.Stop() //allow double escape to quit the app
			}
			app.SetFocus(actions)
			return nil
		}
		return e
	})

	if err := app.Run(); err != nil {
		return err
	}
	return nil
}

func displayAdjustTarget() {
	//update state for event handlers
	selectedAction, _ = actions.GetItemText(actions.GetCurrentItem())
	//update displayed elements in details pane
	details.Clear()
	details.SetTitle(selectedAction)
	notes = tview.NewTextView()
	notes.SetBorder(true)
	notes.SetTitle("Current Values(Radians)")
	notes.SetScrollable(false)
	notes.SetText(fmt.Sprintf("Azi: %v\nAlt: %v\n\nStep size: %.6f", config.Target.Azimuth, config.Target.Altitude, currentMoveSize))
	options := tview.NewTable().SetBorders(false)
	options.SetTitle("Press Buttons to Adjust").SetTitleColor(tcell.ColorForestGreen)
	options.SetBorder(true)
	options.SetCell(0, 1, tview.NewTableCell("(w) UP").SetBackgroundColor(tcell.ColorDarkBlue))
	options.SetCell(1, 0, tview.NewTableCell("(a) LEFT").SetBackgroundColor(tcell.ColorDarkBlue))
	options.SetCell(2, 1, tview.NewTableCell("(s) DOWN").SetBackgroundColor(tcell.ColorDarkBlue))
	options.SetCell(1, 2, tview.NewTableCell("(d) RIGHT").SetBackgroundColor(tcell.ColorDarkBlue))
	options.SetCell(4, 0, tview.NewTableCell("(<) Inc").SetBackgroundColor(tcell.ColorDarkBlue))
	options.SetCell(4, 2, tview.NewTableCell("(>) Dec").SetBackgroundColor(tcell.ColorDarkBlue))

	details.AddItem(notes, 0, 1, false)
	details.AddItem(options, 0, 1, true)
	details.SetInputCapture(adjustTargetEventHandler)

	app.SetFocus(details)
}

func displayLatLong() {
	//update state for event handlers
	selectedAction, _ = actions.GetItemText(actions.GetCurrentItem())
	//update displayed elements in details pane
	details.Clear()
	details.SetTitle(selectedAction)
	options := tview.NewForm().
		AddTextArea("Lat", "", 25, 1, 25, func(t string) {}).
		AddTextArea("Long", "", 25, 1, 25, func(t string) {})
	details.AddItem(options, 0, 1, true)
	details.SetInputCapture(nil)
	app.SetFocus(details)
}

func displayAdjustTime() {
	//update state for event handlers
	selectedAction, _ = actions.GetItemText(actions.GetCurrentItem())
	//update displayed elements in details pane
	details.Clear()
	details.SetTitle(selectedAction)
	options := tview.NewForm().
		AddTextArea("Time", "dd/mm/yyyy", 25, 1, 25, func(t string) {})
	details.AddItem(options, 0, 1, true)
	details.SetInputCapture(nil)
	app.SetFocus(details)
}
