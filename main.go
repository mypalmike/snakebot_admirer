package main

import (
	"context"
	"fmt"
	_ "image/png"
	"log"

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

func makePost(client *mastodon.Client, event *mastodon.UpdateEvent, snakeSpaceGrid [][]SnakeSpace, move string) {
	// Implement post reply logic here
}
