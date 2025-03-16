package game

import "errors"

const (
	BoardSize    = 15
	WinCondition = 5
)

type Player int

const (
	Empty Player = iota
	Black
	White
)

type Board struct {
	Grid         [BoardSize][BoardSize]Player
	CurrentTurn  Player
	MoveHistory  [][2]int
	GameFinished bool
}

func NewBoard() *Board {
	return &Board{
		CurrentTurn: Black,
		MoveHistory: make([][2]int, 0),
	}
}

func (b *Board) PlaceStone(row, col int) error {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return errors.New("position out of bounds")
	}

	if b.Grid[row][col] != Empty {
		return errors.New("position already occupied")
	}

	if b.GameFinished {
		return errors.New("game is already finished")
	}

	b.Grid[row][col] = b.CurrentTurn
	b.MoveHistory = append(b.MoveHistory, [2]int{row, col})

	if b.CheckWin(row, col) {
		b.GameFinished = true
		return nil
	}

	b.CurrentTurn = b.nextPlayer()
	return nil
}

func (b *Board) Undo() error {
	if len(b.MoveHistory) == 0 {
		return errors.New("no moves to undo")
	}

	lastMove := b.MoveHistory[len(b.MoveHistory)-1]
	b.Grid[lastMove[0]][lastMove[1]] = Empty
	b.MoveHistory = b.MoveHistory[:len(b.MoveHistory)-1]
	b.CurrentTurn = b.nextPlayer()
	b.GameFinished = false
	return nil
}

func (b *Board) CheckWin(row, col int) bool {
	directions := [][2]int{
		{1, 0},  // vertical
		{0, 1},  // horizontal
		{1, 1},  // diagonal
		{1, -1}, // anti-diagonal
	}

	player := b.Grid[row][col]
	for _, dir := range directions {
		count := 1
		// Check forward direction
		for i := 1; i < WinCondition; i++ {
			r, c := row+dir[0]*i, col+dir[1]*i
			if !b.isValidPosition(r, c) || b.Grid[r][c] != player {
				break
			}
			count++
		}
		// Check backward direction
		for i := 1; i < WinCondition; i++ {
			r, c := row-dir[0]*i, col-dir[1]*i
			if !b.isValidPosition(r, c) || b.Grid[r][c] != player {
				break
			}
			count++
		}
		if count >= WinCondition {
			return true
		}
	}
	return false
}

func (b *Board) isValidPosition(row, col int) bool {
	return row >= 0 && row < BoardSize && col >= 0 && col < BoardSize
}

func (b *Board) nextPlayer() Player {
	if b.CurrentTurn == Black {
		return White
	}
	return Black
}

func (b *Board) GetCurrentPlayer() Player {
	return b.CurrentTurn
}

func (b *Board) IsGameFinished() bool {
	return b.GameFinished
}
