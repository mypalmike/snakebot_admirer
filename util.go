package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

func testMain() {
	boardWidth := 8
	boardHeight := 5

	// Load image from disk
	imageData, err := loadImageFromDisk("snakebot_test_image.png")
	if err != nil {
		fmt.Println("Failed to load image from disk:", err)
		return
	}

	croppedImageData, err := autocropImage(imageData)
	if err != nil {
		fmt.Println("Failed to autocrop image:", err)
		return
	}

	imageGrid, err := imageToGridImages(croppedImageData, boardWidth, boardHeight)
	if err != nil {
		fmt.Println("Failed to convert image to grid images:", err)
		return
	}

	dumpImageGridToFiles(imageGrid)

	snakeSpaceGrid, err := convertImageGridToSnakeSpaceGrid(imageGrid)
	if err != nil {
		fmt.Println("Failed to convert image grid to SnakeSpace grid:", err)
		return
	}

	gameState, err := convertSnakeSpaceGridToGameState(snakeSpaceGrid)
	if err != nil {
		fmt.Println("Failed to convert SnakeSpace grid to game state:", err)
		return
	}

	fmt.Println("Game state:")
	fmt.Printf("Board width: %v\n", gameState.BoardWidth)
	fmt.Printf("Board height: %v\n", gameState.BoardHeight)
	fmt.Printf("Snake shape: %v\n", gameState.SnakeShape)
	fmt.Printf("Food: %v\n", gameState.Food)
	fmt.Printf("Direction: %v\n", gameState.Direction)

	// Get the best move
	bestMove := determineNextMove(gameState)

	// Respond to the post with the chosen move
	logAnalysis(snakeSpaceGrid, bestMove)
}

func loadImageFromDisk(filename string) (image.Image, error) {
	// Open image file
	imageFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()

	// Decode image
	imageData, _, err := image.Decode(imageFile)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}

func logAnalysis(snakeSpaceGrid [][]SnakeSpace, move string) {
	fmt.Println("SnakeSpace grid:")
	printSnakeSpace(snakeSpaceGrid)
	fmt.Println("Best move:", move)
}

func snakeSpaceGridAsString(snakeSpaceGrid [][]SnakeSpace) string {
	result := ""

	// Top border
	result += "┌"
	for i := 0; i < len(snakeSpaceGrid[0]); i++ {
		result += "─"
	}
	result += "┐\n"

	// Middle
	for iCol, row := range snakeSpaceGrid {
		result += "│"
		for iRow, snakeSpace := range row {
			isEven := (iCol+iRow)%2 == 0
			if snakeSpace.SnakeSlot == Head {
				result += "╋"
			} else if snakeSpace.SnakeSlot == Snake {
				result += lookupAdjacencyUnicode(int64(snakeSpace.Adjacencies))
			} else if snakeSpace.SnakeSlot == Food {
				result += "▖"
			} else if isEven {
				result += "░"
			} else {
				result += "▒"
			}
		}
		result += "│\n"
	}

	// Bottom border
	result += "└"
	for i := 0; i < len(snakeSpaceGrid[0]); i++ {
		result += "─"
	}
	result += "┘"

	return result

}

func lookupAdjacencyUnicode(adjacency int64) string {
	switch adjacency {
	case Up:
		return "╹"
	case Down:
		return "╻"
	case Left:
		return "╸"
	case Right:
		return "╺"
	}

	if adjacency == Up|Down {
		return "║"
	} else if adjacency == Left|Right {
		return "═"
	} else if adjacency == Up|Right {
		return "╚"
	} else if adjacency == Up|Left {
		return "╝"
	} else if adjacency == Down|Right {
		return "╔"
	} else if adjacency == Down|Left {
		return "╗"
	}
	return "┄"
}

func printSnakeSpace(snakeSpaceGrid [][]SnakeSpace) {
	fmt.Print(snakeSpaceGridAsString(snakeSpaceGrid))
}

func dumpImageGridToFiles(imageGrid [][]image.Image) {
	for y, row := range imageGrid {
		for x, image := range row {
			file, err := os.Create(fmt.Sprintf("image_grid_%d_%d.png", x, y))
			if err != nil {
				fmt.Println("Failed to create image file:", err)
				return
			}
			defer file.Close()

			err = png.Encode(file, image)
			if err != nil {
				fmt.Println("Failed to encode image:", err)
				return
			}
		}
	}
}

func countBits(n int64) int {
	count := 0
	for n != 0 {
		count++
		n &= n - 1
	}
	return count
}
