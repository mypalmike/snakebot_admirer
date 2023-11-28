package main

type Position struct {
	X, Y int
}

type GameState struct {
	BoardWidth  int
	BoardHeight int
	SnakeShape  []Position
	Food        Position
	Direction   string
}
