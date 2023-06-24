package main

import (
	"fmt"
	"image"
	"math"
	"net/http"
	"regexp"
	"strconv"
)

func downloadImage(imageURL string) (image.Image, error) {
	// Send HTTP GET request to download the image
	response, err := http.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read and decode the image data.
	// The image.Decode function will automatically detect the image type and decode it.
	imageReader, _, err := image.Decode(response.Body)
	if err != nil {
		return nil, err
	}

	return imageReader, nil
}

func autocropImage(sourceImage image.Image) (image.Image, error) {
	// The image may have alpha boundary pixels, so we need to crop them out
	// to get the actual game board.

	// Iterate over the image top to bottom, left to right, until finding the top left corner.
	// This is the first pixel that is not transparent.
	found := false
	var minX, minY int
	for y := sourceImage.Bounds().Min.Y; y < sourceImage.Bounds().Max.Y; y++ {
		for x := sourceImage.Bounds().Min.X; x < sourceImage.Bounds().Max.X; x++ {
			// Get the color of the current pixel
			_, _, _, a := sourceImage.At(x, y).RGBA()

			// If the pixel is not transparent, we have found the top left corner
			if a != 0 {
				minX = x
				minY = y
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	// Now do the same thing, but starting from the bottom right corner.
	// This is the bottom right corner of the game board.
	found = false
	var maxX, maxY int
	for y := sourceImage.Bounds().Max.Y - 1; y >= sourceImage.Bounds().Min.Y; y-- {
		for x := sourceImage.Bounds().Max.X - 1; x >= sourceImage.Bounds().Min.X; x-- {
			// Get the color of the current pixel
			_, _, _, a := sourceImage.At(x, y).RGBA()

			// If the pixel is not transparent, we have found the bottom right corner
			if a != 0 {
				maxX = x + 1
				maxY = y + 1
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	fmt.Println("Crop image results:")
	fmt.Println("minX:", minX)
	fmt.Println("minY:", minY)
	fmt.Println("maxX:", maxX)
	fmt.Println("maxY:", maxY)

	// Now return a cropped version of the image
	return sourceImage.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(minX, minY, maxX, maxY)), nil
}

func extractBoardDimensions(altText string) (int, int, error) {
	// Regular expression to match the board dimensions
	r := regexp.MustCompile(`(\d+)x(\d+)`)

	// Find the board dimensions in the ALT text
	matches := r.FindStringSubmatch(altText)
	if len(matches) < 3 {
		return 0, 0, fmt.Errorf("failed to extract board dimensions from ALT text")
	}

	// Parse the dimensions as integers
	width, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse board width: %v", err)
	}

	height, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse board height: %v", err)
	}

	return width, height, nil
}

func imageToGridImages(sourceImage image.Image, width int, height int) ([][]image.Image, error) {
	// Determine the dimensions of each image grid
	gridWidth := math.Floor(float64(sourceImage.Bounds().Dx()) / float64(width))
	gridHeight := math.Floor(float64(sourceImage.Bounds().Dy()) / float64(height))

	// Create a 2-dimensional array to store the individual images
	images := make([][]image.Image, height)
	for i := 0; i < height; i++ {
		images[i] = make([]image.Image, width)
	}

	// Split the image into individual grid images
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate the boundaries of the current grid image
			minX := sourceImage.Bounds().Min.X + int(math.Floor(float64(x)*gridWidth))
			maxX := sourceImage.Bounds().Min.X + int(math.Floor(float64(x+1)*gridWidth))
			minY := sourceImage.Bounds().Min.Y + int(math.Floor(float64(y)*gridHeight))
			maxY := sourceImage.Bounds().Min.Y + int(math.Floor(float64(y+1)*gridHeight))

			// Extract the current grid image from the original image
			gridImage := sourceImage.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(minX, minY, maxX, maxY))

			// Store the grid image in the 2-dimensional array
			images[y][x] = gridImage
		}
	}

	return images, nil
}

type SnakeSlot int64

const (
	UndefSlot SnakeSlot = iota
	Empty
	Snake
	Food
	Head
)

type Adjacencies int64

const (
	UndefAdj Adjacencies = 0
	Up                   = 1 << (iota - 1)
	Down
	Left
	Right
)

type SnakeSpace struct {
	SnakeSlot   SnakeSlot
	Adjacencies Adjacencies
}

func imageToSnakeSpace(gridImage image.Image) (SnakeSpace, error) {
	// Define the RGB values for food, snake, and black
	foodColor := color{0x0A, 0x87, 0x54}  // RGB: #0A8754
	snakeColor := color{0x41, 0x6E, 0xD8} // RGB: #416ED8
	blackColor := color{0x00, 0x00, 0x00} // RGB: #000000

	// Initialize the SnakeGrid struct with default values
	snakeSpace := SnakeSpace{
		SnakeSlot:   UndefSlot,
		Adjacencies: UndefAdj,
	}

	// Get the dimensions of the grid image
	imageWidth := gridImage.Bounds().Dx()
	imageHeight := gridImage.Bounds().Dy()

	minX := gridImage.Bounds().Min.X
	minY := gridImage.Bounds().Min.Y

	// Sample the pixels at the midpoints in all four directions
	midpoints := []struct {
		X, Y int
	}{
		{minX + imageWidth/2, minY},                   // Top
		{minX + imageWidth/2, minY + imageHeight - 1}, // Bottom
		{minX, minY + imageHeight/2},                  // Left
		{minX + imageWidth - 1, minY + imageHeight/2}, // Right
	}

	// fmt.Println("Midpoints:")
	// fmt.Println(midpoints)

	adjacencyCount := 0

	// Iterate over the midpoints and check if they match the snake color
	for idx, midpoint := range midpoints {
		pixel := gridImage.At(midpoint.X, midpoint.Y)
		red, green, blue, _ := pixel.RGBA()
		red >>= 8
		green >>= 8
		blue >>= 8

		// Check if the midpoint matches the snake color
		if compareColors(color{red, green, blue}, snakeColor) {
			fmt.Println("Found snake adjacency at midpoint:", midpoint)

			adjacencyCount++

			snakeSpace.SnakeSlot = Snake

			// Determine the direction based on the midpoint
			direction := UndefAdj
			if idx == 0 {
				direction = Up
			} else if idx == 1 {
				direction = Down
			} else if idx == 2 {
				direction = Left
			} else if idx == 3 {
				direction = Right
			}

			// Append the direction to the adjacencies array
			snakeSpace.Adjacencies |= direction

			if adjacencyCount == 0 {
				return snakeSpace, fmt.Errorf("Found no snake adjacencies")
			} else if adjacencyCount > 2 {
				return snakeSpace, fmt.Errorf("Found more than two snake adjacencies")
			}
		}
	}

	if snakeSpace.SnakeSlot == Snake {
		return snakeSpace, nil
	}

	// Iterate over each pixel in the grid image
	bounds := gridImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := gridImage.At(x, y)
			red, green, blue, _ := pixel.RGBA()

			if compareColors(color{red, green, blue}, blackColor) {
				snakeSpace.SnakeSlot = Head
				return snakeSpace, nil
			} else if compareColors(color{red, green, blue}, foodColor) {
				snakeSpace.SnakeSlot = Food
				return snakeSpace, nil
			}
		}
	}

	return snakeSpace, nil
}

func compareColors(c1, c2 color) bool {
	// Compare RGB values with a tolerance for slight differences
	tolerance := 10
	return absDiff(int(c1.r), int(c2.r)) <= tolerance &&
		absDiff(int(c1.g), int(c2.g)) <= tolerance &&
		absDiff(int(c1.b), int(c2.b)) <= tolerance
}

func absDiff(a, b int) int {
	return int(math.Abs(float64(a - b)))
}

type color struct {
	r, g, b uint32
}

func generateSnakeShape(gridData [][]SnakeSpace) ([]Position, error) {
	// Find the head position in the grid
	found := false
	var headPosition Position
	for y := 0; y < len(gridData); y++ {
		for x := 0; x < len(gridData[y]); x++ {
			if gridData[y][x].SnakeSlot == Head {
				headPosition = Position{x, y}
				found = true
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("Could not find snake head")
	}

	// Perform a search to generate the snake array
	var snakeArray []Position
	cameFrom := UndefAdj
	currentPos := headPosition

	length := 0
	for {
		snakeArray = append(snakeArray, currentPos)

		adjacencies := gridData[currentPos.Y][currentPos.X].Adjacencies

		// Remove the adjacency we came from. Note: works correctly for initial case of UndefAdj
		adjacencies &= ^cameFrom

		// Check for tail
		if adjacencies == 0 {
			break
		}

		// Determine the next position based on the adjacencies
		if adjacencies&Down != 0 {
			cameFrom = Up
			currentPos.Y++
		} else if adjacencies&Up != 0 {
			cameFrom = Down
			currentPos.Y--
		} else if adjacencies&Right != 0 {
			cameFrom = Left
			currentPos.X++
		} else if adjacencies&Left != 0 {
			cameFrom = Right
			currentPos.X--
		} else {
			return nil, fmt.Errorf("Could not find a valid adjacency while generating snake array")
		}

		length++

		if length > len(gridData)*len(gridData[0]) {
			return nil, fmt.Errorf("Snake array length exceeded grid size, infinite loop suspected")
		}
	}

	return snakeArray, nil
}

// Helper function to check if a string slice contains a given string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func countOnBits(n int64) int {
	count := 0
	for n != 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

func convertImageGridToSnakeSpaceGrid(imageGrid [][]image.Image) ([][]SnakeSpace, error) {
	// Initialize the snake space grid
	snakeSpaceGrid := make([][]SnakeSpace, len(imageGrid))
	for i := range snakeSpaceGrid {
		snakeSpaceGrid[i] = make([]SnakeSpace, len(imageGrid[i]))
	}

	// Iterate over the image grid and convert each image to a SnakeSpace
	for y := 0; y < len(imageGrid); y++ {
		for x := 0; x < len(imageGrid[y]); x++ {
			snakeSpace, err := imageToSnakeSpace(imageGrid[y][x])
			if err != nil {
				return nil, err
			}
			snakeSpaceGrid[y][x] = snakeSpace
		}
	}

	return snakeSpaceGrid, nil
}

func findFood(snakeSpaceGrid [][]SnakeSpace) (Position, error) {
	for y := 0; y < len(snakeSpaceGrid); y++ {
		for x := 0; x < len(snakeSpaceGrid[y]); x++ {
			if snakeSpaceGrid[y][x].SnakeSlot == Food {
				return Position{x, y}, nil
			}
		}
	}

	return Position{}, fmt.Errorf("Could not find food")
}

func convertSnakeSpaceGridToGameState(snakeSpaceGrid [][]SnakeSpace) (GameState, error) {
	boardHeight := len(snakeSpaceGrid)
	boardWidth := len(snakeSpaceGrid[0])
	snakeShape, err := generateSnakeShape(snakeSpaceGrid)
	if err != nil {
		return GameState{}, err
	}
	food, err := findFood(snakeSpaceGrid)
	if err != nil {
		return GameState{}, err
	}
	direction, err := determineHeadDirection(snakeSpaceGrid)
	if err != nil {
		return GameState{}, err
	}

	return GameState{
		BoardHeight: boardHeight,
		BoardWidth:  boardWidth,
		SnakeShape:  snakeShape,
		Food:        food,
		Direction:   direction,
	}, nil
}

func determineHeadDirection(snakeSpaceGrid [][]SnakeSpace) (string, error) {
	// Find the head position in the grid
	found := false
	var headPosition Position
	for y := 0; y < len(snakeSpaceGrid); y++ {
		for x := 0; x < len(snakeSpaceGrid[y]); x++ {
			if snakeSpaceGrid[y][x].SnakeSlot == Head {
				headPosition = Position{x, y}
				found = true
				break
			}
		}
	}

	if !found {
		return "", fmt.Errorf("Could not find head position")
	}

	// Use adjacency information to determine the direction of the head
	adjacencies := snakeSpaceGrid[headPosition.Y][headPosition.X].Adjacencies
	if adjacencies&Down != 0 {
		return "up", nil
	} else if adjacencies&Up != 0 {
		return "down", nil
	} else if adjacencies&Right != 0 {
		return "left", nil
	} else if adjacencies&Left != 0 {
		return "right", nil
	}

	return "", fmt.Errorf("Could not determine head direction")
}
