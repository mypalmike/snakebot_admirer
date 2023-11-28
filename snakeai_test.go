package main

import "testing"

// Helper function to compare two slices of Positions
func equalPositions(a []Position, b []Position) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, pos := range a {
		if pos != b[idx] {
			return false
		}
	}

	return true
}

// Helper function to compare two game states
func equalGameStates(a GameState, b GameState) bool {
	if a.BoardWidth != b.BoardWidth || a.BoardHeight != b.BoardHeight || a.Direction != b.Direction {
		return false
	}

	if !equalPositions(a.SnakeShape, b.SnakeShape) {
		return false
	}

	if a.Food != b.Food {
		return false
	}

	return true
}

func TestSnakeHead(t *testing.T) {
	// setup game state
	game := GameState{
		SnakeShape: []Position{
			{X: 1, Y: 1},
			{X: 2, Y: 1},
			{X: 2, Y: 2},
		},
		// initialize other fields...
	}

	// call function to test
	result := snakeHead(game)

	// check result
	expected := Position{X: 1, Y: 1}
	if result != expected {
		t.Errorf("snakeHead returned %v, want %v", result, expected)
	}
}

func TestSnakeTail(t *testing.T) {
	// setup game state
	game := GameState{
		SnakeShape: []Position{
			{X: 1, Y: 1},
			{X: 2, Y: 1},
			{X: 2, Y: 2},
		},
		// initialize other fields...
	}

	// call function to test
	result := snakeTail(game)

	// check result
	expected := Position{X: 2, Y: 2}
	if result != expected {
		t.Errorf("snakeTail returned %v, want %v", result, expected)
	}
}

func TestCollidesWithWallFalse(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := collidesWithWall(game)

	// check result
	expected := false
	if result != expected {
		t.Errorf("collidesWithWall returned %v, want %v", result, expected)
	}
}

func TestCollidesWithWallTrue(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  2,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := collidesWithWall(game)

	// check result
	expected := true
	if result != expected {
		t.Errorf("collidesWithWall returned %v, want %v", result, expected)
	}
}

func TestCollidesWithBodyFalse(t *testing.T) {
	// setup game state
	game := GameState{
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := collidesWithBody(game)

	// check result
	expected := false
	if result != expected {
		t.Errorf("collidesWithBody returned %v, want %v", result, expected)
	}
}

func TestCollidesWithBodyTrue(t *testing.T) {
	// setup game state
	game := GameState{
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 2, Y: 2},
			{X: 2, Y: 1}, // Collides here
			{X: 2, Y: 0},
		},
		// initialize other fields...
	}

	// call function to test
	result := collidesWithBody(game)

	// check result
	expected := true
	if result != expected {
		t.Errorf("collidesWithBody returned %v, want %v", result, expected)
	}
}

func TestCollidesWithSomethingFalse(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := collidesWithSomething(game)

	// check result
	expected := false
	if result != expected {
		t.Errorf("collidesWithSomething returned %v, want %v", result, expected)
	}
}

func TestCollidesWithSomethingTrue(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  2,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := collidesWithSomething(game)

	// check result
	expected := true
	if result != expected {
		t.Errorf("collidesWithSomething returned %v, want %v", result, expected)
	}
}

func TestIsTrappedFalse(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := isTrapped(game)

	// check result
	expected := false
	if result != expected {
		t.Errorf("isTrapped returned %v, want %v", result, expected)
	}
}

func TestIsTrappedTrue(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 0, Y: 0}, // Head
			{X: 0, Y: 1},
			{X: 1, Y: 1},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
		},
		// initialize other fields...
	}

	// call function to test
	result := isTrapped(game)

	// check result
	expected := true
	if result != expected {
		t.Errorf("isTrapped returned %v, want %v", result, expected)
	}
}

func TestGetNeighborsMiddle(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 0, Y: 0}, // Head
		},
		// initialize other fields...
	}

	// call function to test
	result := getNeighbors(game, Position{X: 1, Y: 1})

	// check result
	expected := []Position{
		{X: 0, Y: 1},
		{X: 2, Y: 1},
		{X: 1, Y: 0},
		{X: 1, Y: 2},
	}
	if !equalPositions(result, expected) {
		t.Errorf("getNeighbors returned %v, want %v", result, expected)
	}
}

func TestGetNeighborsTopLeft(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 0, Y: 0}, // Head
		},
		// initialize other fields...
	}

	// call function to test
	result := getNeighbors(game, Position{X: 0, Y: 0})

	// check result
	expected := []Position{
		{X: 1, Y: 0},
		{X: 0, Y: 1},
	}
	if !equalPositions(result, expected) {
		t.Errorf("getNeighbors returned %v, want %v", result, expected)
	}
}

func TestGetNeighborsBottomRight(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 10,
		SnakeShape: []Position{
			{X: 2, Y: 9}, // Head
		},
		// initialize other fields...
	}

	// call function to test
	result := getNeighbors(game, Position{X: 2, Y: 9})

	// check result
	expected := []Position{
		{X: 1, Y: 9},
		{X: 2, Y: 8},
	}
	if !equalPositions(result, expected) {
		t.Errorf("getNeighbors returned %v, want %v", result, expected)
	}
}

func TestIsInSnakeMiddleFalse(t *testing.T) {
	// setup game state
	game := GameState{
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		},
		// initialize other fields...
	}

	// call function to test
	result := isInSnakeMiddle(game, Position{X: 1, Y: 2})

	// check result
	expected := false
	if result != expected {
		t.Errorf("isInSnakeMiddle returned %v, want %v", result, expected)
	}
}

func TestIsInSnakeMiddleTrue(t *testing.T) {
	// setup game state
	game := GameState{
		SnakeShape: []Position{
			{X: 2, Y: 1}, // Head
			{X: 1, Y: 1}, // Middle
			{X: 0, Y: 1}, // Tail
		},
		// initialize other fields...
	}

	// call function to test
	result := isInSnakeMiddle(game, Position{X: 1, Y: 1})

	// check result
	expected := true
	if result != expected {
		t.Errorf("isInSnakeMiddle returned %v, want %v", result, expected)
	}
}

func TestMoveInDirection(t *testing.T) {
	// setup game state
	game := GameState{
		BoardWidth:  3,
		BoardHeight: 3,
		SnakeShape: []Position{
			{X: 1, Y: 1}, // Head
			{X: 1, Y: 2}, // Middle
			{X: 1, Y: 3}, // Tail
		},
		Direction: "up",
		// initialize other fields...
	}

	// call function to test
	result := moveInDirection(game, "up")

	// check result
	expected := GameState{
		BoardWidth:  3,
		BoardHeight: 3,
		SnakeShape: []Position{
			{X: 1, Y: 0}, // Head
			{X: 1, Y: 1}, // Middle
			{X: 1, Y: 2}, // Tail
		},
		Direction: "up",
		// initialize other fields...
	}
	if !equalGameStates(result, expected) {
		t.Errorf("simulateMove returned %v, want %v", result, expected)
	}
}
