package main

type Position struct {
	X, Y int
}

// type SnakeShape struct {
// 	Head Position
// 	Body []Position
// 	//	Direction string
// }

type GameState struct {
	BoardWidth  int
	BoardHeight int
	SnakeShape  []Position
	Food        Position
	Direction   string
}
