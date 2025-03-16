package game

import (
	"math"
	"math/rand"
)

type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

type AI struct {
	player     Player
	difficulty Difficulty
}

func NewAI(player Player, difficulty Difficulty) *AI {
	return &AI{
		player:     player,
		difficulty: difficulty,
	}
}

func (ai *AI) MakeMove(board *Board) (int, int) {
	switch ai.difficulty {
	case Easy:
		return ai.makeEasyMove(board)
	case Medium:
		return ai.makeMediumMove(board)
	case Hard:
		return ai.makeHardMove(board)
	default:
		return ai.makeEasyMove(board)
	}
}

// Easy mode: Prevents opponent's winning moves and three-in-a-row threats, prefers valuable positions
func (ai *AI) makeEasyMove(board *Board) (int, int) {
	// 1. Check if AI can win
	if move := ai.findWinningMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 2. Check if need to block opponent's winning move
	if move := ai.findWinningMove(board, ai.getOpponent()); move[0] >= 0 {
		return move[0], move[1]
	}

	// 3. Check opponent's three-in-a-row threats
	if move := ai.findThreatsMove(board); move[0] >= 0 {
		return move[0], move[1]
	}

	// 4. Find the range of existing stones
	minRow, maxRow := BoardSize-1, 0
	minCol, maxCol := BoardSize-1, 0
	hasStones := false

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] != Empty {
				hasStones = true
				if i < minRow {
					minRow = i
				}
				if i > maxRow {
					maxRow = i
				}
				if j < minCol {
					minCol = j
				}
				if j > maxCol {
					maxCol = j
				}
			}
		}
	}

	// If no stones on board, play near center
	if !hasStones {
		center := BoardSize / 2
		return center, center
	}

	// Expand search range, but avoid edges
	minRow = max(2, minRow-2)
	maxRow = min(BoardSize-3, maxRow+2)
	minCol = max(2, minCol-2)
	maxCol = min(BoardSize-3, maxCol+2)

	// 5. Collect possible moves within valid range
	type moveWithWeight struct {
		row    int
		col    int
		weight int
	}
	var moves []moveWithWeight

	// Get last move position
	lastRow, lastCol := BoardSize/2, BoardSize/2
	if len(board.MoveHistory) > 0 {
		lastMove := board.MoveHistory[len(board.MoveHistory)-1]
		lastRow, lastCol = lastMove[0], lastMove[1]
	}

	// Check all empty positions within valid range
	for i := minRow; i <= maxRow; i++ {
		for j := minCol; j <= maxCol; j++ {
			if board.Grid[i][j] == Empty {
				weight := 100

				// Evaluate position value
				weight += ai.evaluatePosition(board, i, j)

				// Adjust weight based on distance to last move
				dist := math.Abs(float64(i-lastRow)) + math.Abs(float64(j-lastCol))
				if dist <= 2 {
					weight += 100 // Very close to last move
				} else if dist <= 4 {
					weight += 50 // Relatively close to last move
				}

				// Adjust weight based on distance to center
				centerDist := math.Abs(float64(i-BoardSize/2)) + math.Abs(float64(j-BoardSize/2))
				if centerDist <= 2 {
					weight += 150 // Close to center
				} else if centerDist <= 4 {
					weight += 80 // Relatively close to center
				}

				// Check for nearby stones
				hasNearbyStones := false
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < BoardSize && nj >= 0 && nj < BoardSize {
							if board.Grid[ni][nj] != Empty {
								hasNearbyStones = true
								weight += 30 // Increase weight for each adjacent stone
							}
						}
					}
				}

				// Significantly reduce weight if no nearby stones
				if !hasNearbyStones {
					weight /= 2
				}

				// Significantly reduce weight for edge positions
				if i <= 1 || i >= BoardSize-2 || j <= 1 || j >= BoardSize-2 {
					weight /= 3
				}

				moves = append(moves, moveWithWeight{i, j, weight})
			}
		}
	}

	if len(moves) > 0 {
		// Calculate total weight
		totalWeight := 0
		for _, move := range moves {
			totalWeight += move.weight
		}

		// Randomly select a position, higher weight means higher chance
		randomWeight := rand.Intn(totalWeight)
		currentWeight := 0
		for _, move := range moves {
			currentWeight += move.weight
			if currentWeight > randomWeight {
				return move.row, move.col
			}
		}

		// If no move selected (shouldn't happen), return the highest weighted move
		bestMove := moves[0]
		for _, move := range moves {
			if move.weight > bestMove.weight {
				bestMove = move
			}
		}
		return bestMove.row, bestMove.col
	}

	// If no suitable position found in valid range, find any empty position
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] == Empty {
				return i, j
			}
		}
	}

	return -1, -1
}

// Find opponent's threats (three-in-a-row, etc.)
func (ai *AI) findThreatsMove(board *Board) [2]int {
	opponent := ai.getOpponent()
	directions := [][2]int{
		{1, 0},  // Vertical
		{0, 1},  // Horizontal
		{1, 1},  // Diagonal
		{1, -1}, // Anti-diagonal
	}

	// Check all empty positions
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] != Empty {
				continue
			}

			// Check each direction
			for _, dir := range directions {
				// Check if this position can block opponent's three-in-a-row
				count := 0
				blocked := 0

				// Forward check
				for k := 1; k < 4; k++ {
					r, c := i+dir[0]*k, j+dir[1]*k
					if !board.isValidPosition(r, c) {
						blocked++
						break
					}
					if board.Grid[r][c] == opponent {
						count++
					} else if board.Grid[r][c] != Empty {
						blocked++
						break
					} else {
						break
					}
				}

				// Backward check
				for k := 1; k < 4; k++ {
					r, c := i-dir[0]*k, j-dir[1]*k
					if !board.isValidPosition(r, c) {
						blocked++
						break
					}
					if board.Grid[r][c] == opponent {
						count++
					} else if board.Grid[r][c] != Empty {
						blocked++
						break
					} else {
						break
					}
				}

				// If found three-in-a-row threat (one end not blocked), block immediately
				if count >= 2 && blocked < 2 {
					return [2]int{i, j}
				}
			}
		}
	}

	return [2]int{-1, -1}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Medium mode: Adds offensive capabilities and strategy to easy mode
func (ai *AI) makeMediumMove(board *Board) (int, int) {
	// 1. Check if AI can win
	if move := ai.findWinningMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 2. Check if need to block opponent's winning move
	if move := ai.findWinningMove(board, ai.getOpponent()); move[0] >= 0 {
		return move[0], move[1]
	}

	// 3. Check if AI can create an open four
	if move := ai.findOpenFourMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 4. Check if opponent can create an open four
	if move := ai.findOpenFourMove(board, ai.getOpponent()); move[0] >= 0 {
		return move[0], move[1]
	}

	// 5. Check if AI can create an open three
	if move := ai.findOpenThreeMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 6. Check opponent's three-in-a-row threats
	if move := ai.findThreatsMove(board); move[0] >= 0 {
		return move[0], move[1]
	}

	// 7. Use evaluation function to find best position
	bestScore := math.MinInt32
	bestMove := [2]int{-1, -1}

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] == Empty {
				score := ai.evaluatePositionMedium(board, i, j)
				if score > bestScore {
					bestScore = score
					bestMove = [2]int{i, j}
				}
			}
		}
	}

	if bestMove[0] >= 0 {
		return bestMove[0], bestMove[1]
	}

	// 8. If no good moves found, use easy mode strategy
	return ai.makeEasyMove(board)
}

// Hard mode: Uses advanced strategies and deep evaluation
func (ai *AI) makeHardMove(board *Board) (int, int) {
	// 1. Check if AI can win
	if move := ai.findWinningMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 2. Check if need to block opponent's winning move
	if move := ai.findWinningMove(board, ai.getOpponent()); move[0] >= 0 {
		return move[0], move[1]
	}

	// 3. Check if AI can create an open four or double-three
	if move := ai.findAdvancedThreatMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 4. Check and block opponent's open four or double-three
	if move := ai.findAdvancedThreatMove(board, ai.getOpponent()); move[0] >= 0 {
		return move[0], move[1]
	}

	// 5. Check if AI can create a single open three
	if move := ai.findOpenThreeMove(board, ai.player); move[0] >= 0 {
		return move[0], move[1]
	}

	// 6. Check if opponent can create a single open three
	if move := ai.findOpenThreeMove(board, ai.getOpponent()); move[0] >= 0 {
		return move[0], move[1]
	}

	// 7. Use advanced evaluation function to find best position
	bestScore := math.MinInt32
	bestMove := [2]int{-1, -1}

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] == Empty {
				score := ai.evaluatePositionHard(board, i, j)
				if score > bestScore {
					bestScore = score
					bestMove = [2]int{i, j}
				}
			}
		}
	}

	if bestMove[0] >= 0 {
		return bestMove[0], bestMove[1]
	}

	// 8. If no good moves found, use medium mode strategy
	return ai.makeMediumMove(board)
}

func (ai *AI) findWinningMove(board *Board, player Player) [2]int {
	// Check all empty positions to see if any can form five in a row
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] == Empty {
				board.Grid[i][j] = player
				if board.CheckWin(i, j) {
					board.Grid[i][j] = Empty
					return [2]int{i, j}
				}
				board.Grid[i][j] = Empty
			}
		}
	}
	return [2]int{-1, -1}
}

// Find positions that can form an open four
func (ai *AI) findOpenFourMove(board *Board, player Player) [2]int {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] != Empty {
				continue
			}
			board.Grid[i][j] = player
			if ai.hasOpenFour(board, i, j) {
				board.Grid[i][j] = Empty
				return [2]int{i, j}
			}
			board.Grid[i][j] = Empty
		}
	}
	return [2]int{-1, -1}
}

// Find positions that can form an open three
func (ai *AI) findOpenThreeMove(board *Board, player Player) [2]int {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] != Empty {
				continue
			}
			board.Grid[i][j] = player
			if ai.hasOpenThree(board, i, j) {
				board.Grid[i][j] = Empty
				return [2]int{i, j}
			}
			board.Grid[i][j] = Empty
		}
	}
	return [2]int{-1, -1}
}

// Find advanced threats (open four or double-three)
func (ai *AI) findAdvancedThreatMove(board *Board, player Player) [2]int {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if board.Grid[i][j] != Empty {
				continue
			}
			board.Grid[i][j] = player

			// Check for open four
			if ai.hasOpenFour(board, i, j) {
				board.Grid[i][j] = Empty
				return [2]int{i, j}
			}

			// Check for double-three
			if ai.hasDoubleThree(board, i, j) {
				board.Grid[i][j] = Empty
				return [2]int{i, j}
			}

			board.Grid[i][j] = Empty
		}
	}
	return [2]int{-1, -1}
}

// Check for double-three formation
func (ai *AI) hasDoubleThree(board *Board, row, col int) bool {
	directions := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	player := board.Grid[row][col]
	threeCount := 0

	for _, dir := range directions {
		count := 1
		space := 0
		blocked := 0

		// Forward check
		for i := 1; i < 4; i++ {
			r, c := row+dir[0]*i, col+dir[1]*i
			if !board.isValidPosition(r, c) {
				blocked++
				break
			}
			if board.Grid[r][c] == player {
				count++
			} else if board.Grid[r][c] == Empty {
				space++
				break
			} else {
				blocked++
				break
			}
		}

		// Backward check
		for i := 1; i < 4; i++ {
			r, c := row-dir[0]*i, col-dir[1]*i
			if !board.isValidPosition(r, c) {
				blocked++
				break
			}
			if board.Grid[r][c] == player {
				count++
			} else if board.Grid[r][c] == Empty {
				space++
				break
			} else {
				blocked++
				break
			}
		}

		// If an open three is formed in this direction
		if count == 3 && space == 2 && blocked == 0 {
			threeCount++
		}
	}

	return threeCount >= 2
}

// Medium difficulty position evaluation
func (ai *AI) evaluatePositionMedium(board *Board, row, col int) int {
	score := ai.evaluatePosition(board, row, col)

	// Check for potential open three or four formations
	board.Grid[row][col] = ai.player
	if ai.hasOpenFour(board, row, col) {
		score += 800
	}
	if ai.hasOpenThree(board, row, col) {
		score += 400
	}
	board.Grid[row][col] = Empty

	// Check for blocking opponent's open three or four
	opponent := ai.getOpponent()
	board.Grid[row][col] = opponent
	if ai.hasOpenFour(board, row, col) {
		score += 700
	}
	if ai.hasOpenThree(board, row, col) {
		score += 300
	}
	board.Grid[row][col] = Empty

	return score
}

// Hard difficulty position evaluation
func (ai *AI) evaluatePositionHard(board *Board, row, col int) int {
	score := ai.evaluatePosition(board, row, col)

	// Check offensive potential
	board.Grid[row][col] = ai.player
	if ai.hasOpenFour(board, row, col) {
		score += 1200
	}
	if ai.hasDoubleThree(board, row, col) {
		score += 1000
	}
	if ai.hasOpenThree(board, row, col) {
		score += 600
	}
	board.Grid[row][col] = Empty

	// Check defensive needs
	opponent := ai.getOpponent()
	board.Grid[row][col] = opponent
	if ai.hasOpenFour(board, row, col) {
		score += 1000
	}
	if ai.hasDoubleThree(board, row, col) {
		score += 800
	}
	if ai.hasOpenThree(board, row, col) {
		score += 500
	}
	board.Grid[row][col] = Empty

	// Consider strategic value
	// 1. Center proximity value
	centerDist := math.Abs(float64(row-BoardSize/2)) + math.Abs(float64(col-BoardSize/2))
	score -= int(centerDist * 15)

	// 2. Value proximity to existing stones
	nearbyStones := 0
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			r, c := row+i, col+j
			if r >= 0 && r < BoardSize && c >= 0 && c < BoardSize {
				if board.Grid[r][c] != Empty {
					dist := math.Abs(float64(i)) + math.Abs(float64(j))
					if dist <= 1 {
						nearbyStones += 3
					} else {
						nearbyStones++
					}
				}
			}
		}
	}
	score += nearbyStones * 10

	// 3. Reduce value for edge positions
	if row <= 1 || row >= BoardSize-2 || col <= 1 || col >= BoardSize-2 {
		score /= 2
	}

	return score
}

func (ai *AI) evaluatePosition(board *Board, row, col int) int {
	score := 0
	directions := [][2]int{
		{1, 0},  // Vertical
		{0, 1},  // Horizontal
		{1, 1},  // Diagonal
		{1, -1}, // Anti-diagonal
	}

	// Check for winning move
	board.Grid[row][col] = ai.player
	if board.CheckWin(row, col) {
		board.Grid[row][col] = Empty
		return 10000
	}
	board.Grid[row][col] = Empty

	// Check for blocking opponent's win
	opponent := ai.getOpponent()
	board.Grid[row][col] = opponent
	if board.CheckWin(row, col) {
		board.Grid[row][col] = Empty
		return 9000
	}
	board.Grid[row][col] = Empty

	// Evaluate each direction
	for _, dir := range directions {
		score += ai.evaluateDirection(board, row, col, dir[0], dir[1])
	}

	// Prefer positions closer to center
	centerDist := math.Abs(float64(row-BoardSize/2)) + math.Abs(float64(col-BoardSize/2))
	score -= int(centerDist * 10)

	// Prefer positions closer to last move
	if len(board.MoveHistory) > 0 {
		lastMove := board.MoveHistory[len(board.MoveHistory)-1]
		lastDist := math.Abs(float64(row-lastMove[0])) + math.Abs(float64(col-lastMove[1]))
		score -= int(lastDist * 5)
	}

	return score
}

func (ai *AI) evaluateDirection(board *Board, row, col, dRow, dCol int) int {
	score := 0
	myCount := 0
	oppCount := 0
	empty := 0
	maxMySeq := 0  // Maximum consecutive own stones
	maxOppSeq := 0 // Maximum consecutive opponent stones
	currentMySeq := 0
	currentOppSeq := 0

	// Check 4 positions in both directions
	for i := -4; i <= 4; i++ {
		r, c := row+dRow*i, col+dCol*i
		if r < 0 || r >= BoardSize || c < 0 || c >= BoardSize {
			continue
		}

		current := board.Grid[r][c]
		if current == ai.player {
			myCount++
			currentMySeq++
			currentOppSeq = 0
			if currentMySeq > maxMySeq {
				maxMySeq = currentMySeq
			}
		} else if current == Empty {
			empty++
			currentMySeq = 0
			currentOppSeq = 0
		} else {
			oppCount++
			currentOppSeq++
			currentMySeq = 0
			if currentOppSeq > maxOppSeq {
				maxOppSeq = currentOppSeq
			}
		}
	}

	// Scoring rules
	if maxMySeq >= 4 {
		score += 2000
	} else if maxMySeq == 3 && empty >= 2 {
		score += 1000
	} else if maxMySeq == 2 && empty >= 3 {
		score += 100
	}

	// Defensive scoring
	if maxOppSeq >= 3 {
		score += 1500
	} else if maxOppSeq == 2 && empty >= 3 {
		score += 200
	}

	// Consider total stone count
	score += myCount * 10
	score += empty * 2

	return score
}

func (ai *AI) hasOpenFour(board *Board, row, col int) bool {
	directions := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	player := board.Grid[row][col]

	for _, dir := range directions {
		count := 1
		space := 0

		// Forward check
		for i := 1; i < 5; i++ {
			r, c := row+dir[0]*i, col+dir[1]*i
			if !board.isValidPosition(r, c) {
				break
			}
			if board.Grid[r][c] == player {
				count++
			} else if board.Grid[r][c] == Empty {
				space++
				break
			} else {
				break
			}
		}

		// Backward check
		for i := 1; i < 5; i++ {
			r, c := row-dir[0]*i, col-dir[1]*i
			if !board.isValidPosition(r, c) {
				break
			}
			if board.Grid[r][c] == player {
				count++
			} else if board.Grid[r][c] == Empty {
				space++
				break
			} else {
				break
			}
		}

		if count == 4 && space == 2 {
			return true
		}
	}
	return false
}

func (ai *AI) hasOpenThree(board *Board, row, col int) bool {
	directions := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	player := board.Grid[row][col]

	for _, dir := range directions {
		count := 1
		space := 0

		// Forward check
		for i := 1; i < 4; i++ {
			r, c := row+dir[0]*i, col+dir[1]*i
			if !board.isValidPosition(r, c) {
				break
			}
			if board.Grid[r][c] == player {
				count++
			} else if board.Grid[r][c] == Empty {
				space++
				break
			} else {
				break
			}
		}

		// Backward check
		for i := 1; i < 4; i++ {
			r, c := row-dir[0]*i, col-dir[1]*i
			if !board.isValidPosition(r, c) {
				break
			}
			if board.Grid[r][c] == player {
				count++
			} else if board.Grid[r][c] == Empty {
				space++
				break
			} else {
				break
			}
		}

		if count == 3 && space == 2 {
			return true
		}
	}
	return false
}

func (ai *AI) getOpponent() Player {
	if ai.player == Black {
		return White
	}
	return Black
}
