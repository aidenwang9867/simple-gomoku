package ui

import (
	"fmt"
	"image/color"
	"os/exec"
	"runtime"
	"time"

	"simple-gomoku/game"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Click area widget, only handles click events
type ClickArea struct {
	widget.BaseWidget
	onTapped func()
}

func NewClickArea(onTapped func()) *ClickArea {
	area := &ClickArea{
		onTapped: onTapped,
	}
	area.ExtendBaseWidget(area)
	return area
}

func (c *ClickArea) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (c *ClickArea) Tapped(_ *fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}

type GameWindow struct {
	window         fyne.Window
	board          *game.Board
	ai             *game.AI
	stones         [][]*canvas.Circle // Store stone displays
	clickAreas     [][]*ClickArea     // Store click areas
	statusLabel    *widget.Label
	isProcessing   bool
	boardContainer *fyne.Container
	lastMoveMarker *fyne.Container // Last move marker
}

func NewGameWindow(window fyne.Window) *GameWindow {
	gw := &GameWindow{
		window: window,
		board:  game.NewBoard(),
		ai:     game.NewAI(game.White, game.Easy), // Create a default AI
	}

	// Initialize UI first to ensure board rendering
	gw.initializeUI()

	// Ensure UI is fully rendered
	gw.window.Canvas().Content().Refresh()

	// Then show difficulty selection dialog
	gw.showDifficultyDialog()
	return gw
}

func (gw *GameWindow) showDifficultyDialog() {
	difficultySelect := widget.NewSelect([]string{"Easy", "Medium", "Hard"}, func(selected string) {
		var difficulty game.Difficulty
		switch selected {
		case "Easy":
			difficulty = game.Easy
		case "Medium":
			difficulty = game.Medium
		case "Hard":
			difficulty = game.Hard
		default:
			difficulty = game.Easy
		}
		gw.ai = game.NewAI(game.White, difficulty)
		gw.board = game.NewBoard() // Reset board
		gw.updateBoard()           // Update UI
	})
	difficultySelect.SetSelected("Easy") // Default to Easy difficulty

	content := container.NewVBox(
		widget.NewLabel("Select AI Difficulty:"),
		difficultySelect,
	)

	dialog := dialog.NewCustom(
		"Game Settings",
		"Start Game",
		content,
		gw.window,
	)

	dialog.Show()
}

func (gw *GameWindow) initializeUI() {
	const (
		cellSize  = float32(40) // Cell size
		padding   = float32(30) // Add padding to ensure complete board display
		stoneSize = float32(32) // Stone size
	)

	boardSize := float32(game.BoardSize-1) * cellSize // Actual board size (distance between lines)
	totalSize := boardSize + padding*2                // Total size (including padding)

	// Initialize storage
	gw.stones = make([][]*canvas.Circle, game.BoardSize)
	gw.clickAreas = make([][]*ClickArea, game.BoardSize)
	gw.boardContainer = container.NewWithoutLayout()

	// 1. Create background
	background := canvas.NewRectangle(color.RGBA{R: 255, G: 223, B: 176, A: 255})
	background.Resize(fyne.NewSize(totalSize, totalSize))
	background.Move(fyne.NewPos(0, 0))
	gw.boardContainer.Add(background)

	// 2. Create grid lines
	for i := 0; i < game.BoardSize; i++ {
		// Horizontal line
		hLine := canvas.NewLine(color.Black)
		hLine.StrokeWidth = 1
		hLine.Move(fyne.NewPos(padding, padding+float32(i)*cellSize))
		hLine.Resize(fyne.NewSize(boardSize, 1))
		gw.boardContainer.Add(hLine)

		// Vertical line
		vLine := canvas.NewLine(color.Black)
		vLine.StrokeWidth = 1
		vLine.Move(fyne.NewPos(padding+float32(i)*cellSize, padding))
		vLine.Resize(fyne.NewSize(1, boardSize))
		gw.boardContainer.Add(vLine)
	}

	// 3. Create stones and click areas
	for i := 0; i < game.BoardSize; i++ {
		gw.stones[i] = make([]*canvas.Circle, game.BoardSize)
		gw.clickAreas[i] = make([]*ClickArea, game.BoardSize)

		for j := 0; j < game.BoardSize; j++ {
			// Create stone (initially transparent)
			stone := canvas.NewCircle(color.Transparent)
			stone.Resize(fyne.NewSize(stoneSize, stoneSize))
			stone.Move(fyne.NewPos(
				padding+float32(j)*cellSize-stoneSize/2,
				padding+float32(i)*cellSize-stoneSize/2,
			))
			gw.stones[i][j] = stone
			gw.boardContainer.Add(stone)

			// Create click area
			clickArea := NewClickArea(func(row, col int) func() {
				return func() {
					gw.handleClick(row, col)
				}
			}(i, j))

			// Set click area size to half of cell size to ensure clicks only near intersections
			clickSize := cellSize * 0.5
			clickArea.Resize(fyne.NewSize(clickSize, clickSize))
			clickArea.Move(fyne.NewPos(
				padding+float32(j)*cellSize-clickSize/2,
				padding+float32(i)*cellSize-clickSize/2,
			))

			gw.clickAreas[i][j] = clickArea
			gw.boardContainer.Add(clickArea)
		}
	}

	// 4. Create control panel
	gw.statusLabel = widget.NewLabel("Black's turn")
	undoButton := widget.NewButton("Undo", func() {
		if gw.isProcessing || gw.board.IsGameFinished() {
			return
		}
		gw.isProcessing = true
		if err := gw.board.Undo(); err == nil {
			if gw.board.GetCurrentPlayer() == game.White {
				gw.board.Undo()
			}
			gw.updateBoard()
			gw.updateStatus()
		}
		gw.isProcessing = false
	})

	newGameButton := widget.NewButton("New Game", func() {
		gw.board = game.NewBoard()
		gw.showDifficultyDialog()
	})

	controls := container.NewHBox(gw.statusLabel, undoButton, newGameButton)
	mainContainer := container.NewBorder(nil, controls, nil, nil, gw.boardContainer)

	// 5. Set window content and size
	gw.window.SetContent(mainContainer)
	gw.window.Resize(fyne.NewSize(totalSize, totalSize+50))
}

func playSystemSound() {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("afplay", "/System/Library/Sounds/Tink.aiff").Run()
	case "linux":
		exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/bell.oga").Run()
	case "windows":
		exec.Command("powershell", "[console]::beep(2000,100)").Run()
	}
}

func (gw *GameWindow) handleClick(row, col int) {
	if gw.isProcessing || gw.board.IsGameFinished() {
		return
	}
	gw.isProcessing = true

	if row < 0 || row >= game.BoardSize || col < 0 || col >= game.BoardSize {
		gw.isProcessing = false
		return
	}

	if gw.board.GetCurrentPlayer() != game.Black {
		gw.isProcessing = false
		return
	}

	if err := gw.board.PlaceStone(row, col); err == nil {
		// Human player stone animation
		stone := gw.stones[row][col]
		stone.FillColor = color.Black
		stone.Refresh()
		gw.updateLastMoveMarker(row, col)
		gw.updateStatus()

		// Play system sound
		go playSystemSound()

		if gw.board.IsGameFinished() {
			gw.showGameOver("Black")
			gw.isProcessing = false
			return
		}

		// AI's turn (with delay)
		go func() {
			time.Sleep(300 * time.Millisecond)

			aiRow, aiCol := gw.ai.MakeMove(gw.board)
			if aiRow >= 0 && aiCol >= 0 {
				// Update UI in main thread
				gw.window.Canvas().Content().Refresh()
				gw.board.PlaceStone(aiRow, aiCol)

				// AI stone animation
				stone := gw.stones[aiRow][aiCol]
				stone.FillColor = color.White
				stone.Refresh()
				gw.updateLastMoveMarker(aiRow, aiCol)
				gw.updateStatus()

				// Play system sound
				go playSystemSound()

				if gw.board.IsGameFinished() {
					gw.showGameOver("White")
				}
			}
			gw.isProcessing = false
		}()
	} else {
		gw.isProcessing = false
	}
}

func (gw *GameWindow) updateBoard() {
	for i := 0; i < game.BoardSize; i++ {
		for j := 0; j < game.BoardSize; j++ {
			switch gw.board.Grid[i][j] {
			case game.Black:
				gw.stones[i][j].FillColor = color.Black
			case game.White:
				gw.stones[i][j].FillColor = color.White
			default:
				gw.stones[i][j].FillColor = color.Transparent
			}
			gw.stones[i][j].Refresh()
		}
	}
}

func (gw *GameWindow) updateStatus() {
	if gw.board.IsGameFinished() {
		gw.statusLabel.SetText("Game Over")
	} else {
		gw.statusLabel.SetText(fmt.Sprintf("%s's turn", gw.getPlayerText(gw.board.GetCurrentPlayer())))
	}
}

func (gw *GameWindow) showGameOver(winner string) {
	content := widget.NewLabel(fmt.Sprintf("Game Over! %s wins!", winner))
	dialog := dialog.NewCustomConfirm(
		"Game Over",
		"New Game",
		"Return to Board",
		content,
		func(ok bool) {
			if ok {
				gw.board = game.NewBoard()
				gw.showDifficultyDialog()
			}
		},
		gw.window,
	)
	dialog.Show()
}

func (gw *GameWindow) getPlayerText(player game.Player) string {
	if player == game.Black {
		return "Black"
	}
	return "White"
}

func (gw *GameWindow) updateLastMoveMarker(row, col int) {
	if gw.lastMoveMarker != nil {
		gw.boardContainer.Remove(gw.lastMoveMarker)
	}

	const (
		cellSize   = float32(40) // Cell size
		padding    = float32(30) // Padding
		markerSize = float32(10) // Marker size
	)

	// Create marker container
	markerContainer := container.NewWithoutLayout()

	// Create horizontal marker line
	hLine := canvas.NewLine(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	hLine.StrokeWidth = 2
	hLine.Resize(fyne.NewSize(markerSize, 1))
	hLine.Move(fyne.NewPos(
		padding+float32(col)*cellSize-markerSize/2,
		padding+float32(row)*cellSize,
	))
	markerContainer.Add(hLine)

	// Create vertical marker line
	vLine := canvas.NewLine(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	vLine.StrokeWidth = 2
	vLine.Resize(fyne.NewSize(1, markerSize))
	vLine.Move(fyne.NewPos(
		padding+float32(col)*cellSize,
		padding+float32(row)*cellSize-markerSize/2,
	))
	markerContainer.Add(vLine)

	gw.lastMoveMarker = markerContainer
	gw.boardContainer.Add(markerContainer)
	markerContainer.Refresh()
}

func (gw *GameWindow) Show() {
	gw.window.Show()
}
