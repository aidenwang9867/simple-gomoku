package main

import (
	"simple-gomoku/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Gomoku Game")
	window.Resize(fyne.NewSize(600, 600))

	game := ui.NewGameWindow(window)
	game.Show()

	window.ShowAndRun()
}
