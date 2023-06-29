package main

import "fmt"

func determineNextMove(game GameState) string {
	possibleMoves := []string{"up", "down", "left", "right"}

	// Evaluate each possible move and choose the best one
	bestMove := ""
	bestScore := -1

	for _, move := range possibleMoves {
		score := evaluateMove(game, move)

		fmt.Println("Move:", move, "Score:", score)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

func gameOver(game GameState) bool {
	// Check if snake collides with the wall or itself
	return collidesWithWall(game) || collidesWithBody(game)
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

func isTrapped(game GameState) bool {
	visited := make(map[Position]bool)
	return !hasPath(snakeHead(game), snakeTail(game), visited, game.BoardWidth, game.BoardHeight)
}

func hasPath(start, target Position, visited map[Position]bool, boardWidth, boardHeight int) bool {
	visited[start] = true

	// Check if we reached the target position
	if start == target {
		return true
	}

	// Check all possible neighbor positions
	neighbors := getNeighbors(start, boardWidth, boardHeight)
	for _, neighbor := range neighbors {
		if !visited[neighbor] {
			if hasPath(neighbor, target, visited, boardWidth, boardHeight) {
				return true
			}
		}
	}

	return false
}

func getNeighbors(pos Position, boardWidth, boardHeight int) []Position {
	neighbors := []Position{
		{pos.X - 1, pos.Y}, // Left
		{pos.X + 1, pos.Y}, // Right
		{pos.X, pos.Y - 1}, // Up
		{pos.X, pos.Y + 1}, // Down
	}

	validNeighbors := make([]Position, 0, 4)

	for _, neighbor := range neighbors {
		if neighbor.X >= 0 && neighbor.X < boardWidth && neighbor.Y >= 0 && neighbor.Y < boardHeight {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}

	return validNeighbors
}

func evaluateMove(game GameState, move string) int {
	// Simulate the move and evaluate the resulting game state
	simulatedGame := simulateMove(game, move)

	// Assign scores based on different criteria
	score := 0

	if gameOver(simulatedGame) {
		// Avoiding collisions
		score -= 2 * game.BoardWidth * game.BoardHeight
	} else if isTrapped(simulatedGame) {
		// Avoid getting trapped
		score -= game.BoardWidth * game.BoardHeight
	}

	// Distance to the food
	distance := calculateManhattanDistance(snakeHead(simulatedGame), simulatedGame.Food)
	score += game.BoardWidth*game.BoardHeight - distance

	return score
}

func simulateMove(game GameState, move string) GameState {
	// Create a copy of the game to simulate the move
	simulatedGame := game

	// Update the snake's direction
	simulatedGame.Direction = move

	// Move the snake
	moveSnake(&simulatedGame)

	return simulatedGame
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

func updateGameState(game *GameState, move string) {
	// Update the game state based on the chosen move
	game.Direction = move
	moveSnake(game)
}

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
