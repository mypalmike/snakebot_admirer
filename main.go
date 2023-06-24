package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	_ "image/png"
	"log"
	"os"

	"github.com/mattn/go-mastodon"
)

func main() {
	testMain()

	return

	// Create a new Mastodon client
	client := mastodon.NewClient(&mastodon.Config{
		Server:       "https://botsin.space",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		AccessToken:  "your-access-token",
	})

	// Create a channel to receive updates from the user's home timeline
	stream, err := client.StreamingUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Start listening to the Mastodon stream
	for event := range stream {
		switch event := event.(type) {
		case *mastodon.UpdateEvent:
			// Check if the update is from the snake bot
			if event.Status.Account.Acct == "snake_game@botsin.space" {
				// Extract the image attachment
				if len(event.Status.MediaAttachments) > 0 && event.Status.MediaAttachments[0].Type == "image" {
					imageData, err := downloadImage(event.Status.MediaAttachments[0].URL)
					if err != nil {
						fmt.Println("Failed to download image:", err)
						return
					}

					croppedImageData, err := autocropImage(imageData)
					if err != nil {
						fmt.Println("Failed to autocrop image:", err)
						return
					}

					// Extract board dimensions from ALT data
					boardWidth, boardHeight, err := extractBoardDimensions(event.Status.MediaAttachments[0].Description)
					if err != nil {
						fmt.Println("Failed to extract board dimensions:", err)
						return
					}

					imageGrid, err := imageToGridImages(croppedImageData, boardWidth, boardHeight)
					if err != nil {
						fmt.Println("Failed to convert image to grid images:", err)
						return
					}

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

					// rawGrid, err := convertImageToRawGrid(imageData, boardWidth, boardHeight)
					// if err != nil {
					// 	fmt.Println("Failed to convert image to raw grid:", err)
					// 	return
					// }

					// // Convert the image into a game state
					// gameState, err := convertImageToGameState(imageData, boardWidth, boardHeight)
					// if err != nil {
					// 	fmt.Println("Failed to convert image to game state:", err)
					// 	return
					// }

					// Get the best move
					bestMove := determineNextMove(gameState)

					// Respond to the post with the chosen move
					makePost(client, event, snakeSpaceGrid, bestMove)
				}
			}
		}
	}
}

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

	logAnalysis(snakeSpaceGrid, "none")

	gameState, err := convertSnakeSpaceGridToGameState(snakeSpaceGrid)
	if err != nil {
		fmt.Println("Failed to convert SnakeSpace grid to game state:", err)
		return
	}

	// Get the best move
	bestMove := determineNextMove(gameState)

	// Respond to the post with the chosen move
	logAnalysis(snakeSpaceGrid, bestMove)
}

// func convertImageToGameState(imageData []byte, width, height int) (snakeai.Game, error) {
// 	// Implement image to game state conversion logic here
// }

func makePost(client *mastodon.Client, event *mastodon.UpdateEvent, snakeSpaceGrid [][]SnakeSpace, move string) {
	// Implement post reply logic here
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
	for _, row := range snakeSpaceGrid {
		for _, cell := range row {
			printSnakeSpace(cell)
			//			fmt.Printf("%v ", cell)
		}
		fmt.Println()
	}
	fmt.Println("Best move:", move)
}

func lookupAdjacencyUnicode(adjacency int64) string {
	switch adjacency {
	case Up:
		return "↑"
	case Down:
		return "↓"
	case Left:
		return "←"
	case Right:
		return "→"
	}

	if adjacency == Up|Down {
		return "│"
	} else if adjacency == Left|Right {
		return "─"
	} else if adjacency == Up|Right {
		return "└"
	} else if adjacency == Up|Left {
		return "┘"
	} else if adjacency == Down|Right {
		return "┌"
	} else if adjacency == Down|Left {
		return "┐"
	}
	return "?"
}

func printSnakeSpace(snakeSpace SnakeSpace) {
	if snakeSpace.SnakeSlot == Head {
		fmt.Print("H")
	} else if snakeSpace.SnakeSlot == Snake {
		fmt.Print(lookupAdjacencyUnicode(int64(snakeSpace.Adjacencies)))
	} else if snakeSpace.SnakeSlot == Food {
		fmt.Print("F")
	} else {
		fmt.Print(".")
	}
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
