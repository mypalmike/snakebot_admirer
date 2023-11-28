package main

import (
	"fmt"
	"math/rand"
)

func calculateManhattanDistance(p1, p2 Position) int {
	// Calculate the Manhattan distance between two positions
	return abs(p1.X-p2.X) + abs(p1.Y-p2.Y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func snakeHead(game GameState) Position {
	return game.SnakeShape[0]
}

func snakeTail(game GameState) Position {
	return game.SnakeShape[len(game.SnakeShape)-1]
}

func collidesWithWall(game GameState) bool {
	head := snakeHead(game)
	return head.X < 0 || head.X >= game.BoardWidth || head.Y < 0 || head.Y >= game.BoardHeight
}

func collidesWithBody(game GameState) bool {
	head := snakeHead(game)
	for idx, bodyPart := range game.SnakeShape {
		if idx != 0 && head.X == bodyPart.X && head.Y == bodyPart.Y {
			return true
		}
	}
	return false
}

func collidesWithSomething(game GameState) bool {
	// Check if snake collides with the wall or itself
	return collidesWithWall(game) || collidesWithBody(game)
}

func isTrapped(game GameState) bool {
	head := snakeHead(game)
	tail := snakeTail(game)
	visited := make(map[Position]bool)
	return !dfsReachable(game, head, tail, visited)
}

func dfsReachable(game GameState, current, target Position, visited map[Position]bool) bool {
	if current == target {
		return true
	}

	visited[current] = true

	for _, neighbor := range getNeighbors(game, current) {
		if !isInSnakeMiddle(game, neighbor) && !visited[neighbor] {
			if dfsReachable(game, neighbor, target, visited) {
				return true
			}
		}
	}

	return false
}

func getNeighbors(game GameState, pos Position) []Position {
	neighbors := []Position{
		{pos.X - 1, pos.Y},
		{pos.X + 1, pos.Y},
		{pos.X, pos.Y - 1},
		{pos.X, pos.Y + 1},
	}

	validNeighbors := []Position{}
	for _, neighbor := range neighbors {
		if neighbor.X >= 0 && neighbor.X < game.BoardWidth && neighbor.Y >= 0 && neighbor.Y < game.BoardHeight {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}

	return validNeighbors
}

// Returns true if the given position is in the middle of the snake's body
// i.e. not the head or tail
func isInSnakeMiddle(game GameState, pos Position) bool {
	for i := 1; i < len(game.SnakeShape)-1; i++ {
		if game.SnakeShape[i] == pos {
			return true
		}
	}
	return false
}

func evaluateMove(game GameState, move string) (int, bool) {
	// Simulate the move and evaluate the resulting game state
	simulatedGame := moveInDirection(game, move)

	// Assign scores based on different criteria
	score := 0
	isDeadly := false

	if collidesWithSomething(simulatedGame) {
		// Avoiding collisions
		score -= 2 * game.BoardWidth * game.BoardHeight
		isDeadly = true
	} else if isTrapped(simulatedGame) {
		// Avoid getting trapped
		score -= game.BoardWidth * game.BoardHeight
		isDeadly = true
	}

	// Distance to the food
	distance := calculateManhattanDistance(snakeHead(simulatedGame), simulatedGame.Food)
	score += game.BoardWidth*game.BoardHeight - distance

	return score, isDeadly
}

func moveInDirection(game GameState, move string) GameState {
	// Create a copy of the game to simulate the move
	newGameState := game

	// Update the snake's direction
	newGameState.Direction = move

	// Move the snake
	moveSnake(&newGameState)

	return newGameState
}

func moveSnake(game *GameState) {
	// Move the snake by updating the position of the head and shifting the body
	head := snakeHead(*game)
	newHead := head

	// Update the position of the head based on the current direction
	switch game.Direction {
	case "up":
		newHead = Position{X: head.X, Y: head.Y - 1}
	case "down":
		newHead = Position{X: head.X, Y: head.Y + 1}
	case "left":
		newHead = Position{X: head.X - 1, Y: head.Y}
	case "right":
		newHead = Position{X: head.X + 1, Y: head.Y}
	}

	// Update the position of the head, removing the tail
	game.SnakeShape = append([]Position{newHead}, game.SnakeShape...)
	if len(game.SnakeShape) > 1 {
		game.SnakeShape = game.SnakeShape[:len(game.SnakeShape)-1]
	}
}

func determineNextMove(game GameState) (string, bool) {
	possibleMoves := []string{"up", "down", "left", "right"}

	// Evaluate each possible move and choose the best one
	bestMove := ""
	bestScore := -1
	mustTurn := false

	for _, move := range possibleMoves {
		score, isDeadly := evaluateMove(game, move)

		fmt.Println("Move:", move, "Score:", score)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if isDeadly {
			if move == game.Direction {
				mustTurn = true
			}
		}

	}

	return bestMove, mustTurn
}

func isValidUnvisitedMove(game GameState, move string, visited map[Position]bool) bool {
	// Check if the move collides with the wall or itself.
	// Note: Moving into the tail is allowed, which is why we check for collisions with the updated snake shape.
	gameStateAfterMove := moveInDirection(game, move)
	if collidesWithSomething(gameStateAfterMove) {
		return false
	}

	// Check if the move revisits a position in the current walk
	if visited[snakeHead(gameStateAfterMove)] {
		return false
	}

	return true
}

func randomValidUnvisitedMove(game GameState, visited map[Position]bool) string {
	// Get all valid moves
	possibleMoves := []string{"up", "down", "left", "right"}
	validMoves := []string{}

	for _, move := range possibleMoves {
		if isValidUnvisitedMove(game, move, visited) {
			validMoves = append(validMoves, move)
		}
	}

	// Possibly no valid moves. This happens when the snake is trapped.
	if len(validMoves) == 0 {
		return ""
	}

	// Choose a random valid move
	return validMoves[rand.Intn(len(validMoves))]
}

// Generate one random valid walk to the food. Avoid collisions, getting trapped, or revisiting
// any spot in the current walk.
func randomWalkToFood(game GameState) []string {
	// Try 100 times to generate a valid walk
	for i := 0; i < 100; i++ {
		// Generate a random walk to the food
		walk := []string{}

		// Keep track of visited positions to avoid getting trapped
		visited := make(map[Position]bool)

		head := snakeHead(game)

		// Keep walking until we reach the food or give up after 25 times getting trapped
		attemptsLeft := 25
		for head != game.Food && attemptsLeft > 0 {
			attemptsLeft -= 1

			// Get the next move
			move := randomValidUnvisitedMove(game, visited)
			if move == "" {
				break
			}

			walk = append(walk, move)

			// // Update the walk and score
			// walk = append(walk, prevHead)

			// Update the current position
			newGame := moveInDirection(game, move)
			head = snakeHead(newGame)
		}

		return walk
	}

	return []string{}
}

func runMonteCarloSimulation(game GameState, numSimulations, depth int) string {
	// bestScore := 0
	// bestWalk := []Position{}

	// for i := 0; i < numSimulations; i++ {
	// 	// Recursively simulate random walks to the food
	// 	walk, score := simulateRandomWalkToFood(game, depth)

	// }

	return "Not implemented"
}

// func updateGameState(game *GameState, move string) {
// 	// Update the game state based on the chosen move
// 	game.Direction = move
// 	moveSnake(game)
// }
