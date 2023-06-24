package main

import (
	"context"
	"fmt"
	_ "image/png"
	"log"
	"os"

	"github.com/mattn/go-mastodon"
)

func main() {
	mastodonServer := os.Getenv("MASTODON_SERVER")
	clientKey := os.Getenv("CLIENT_KEY")
	clientSecret := os.Getenv("CLIENT_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")

	client := mastodon.NewClient(&mastodon.Config{
		Server:       mastodonServer,
		ClientID:     clientKey,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
	})

	fmt.Println("Connected to Mastodon server")

	// Create a channel to receive updates from the user's home timeline
	stream, err := client.StreamingUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening to Mastodon stream")

	// Start listening to the Mastodon stream
	for event := range stream {
		fmt.Println("Received event:", event)
		switch event := event.(type) {
		case *mastodon.UpdateEvent:
			fmt.Println("-> and it's an update event from " + event.Status.Account.Acct)
			// Check if the update is from the snake bot
			if event.Status.Account.Acct == "snake_game@botsin.space" || event.Status.Account.Acct == "snake_game" {
				fmt.Println("-> and it's from the snake bot")
				// Extract the image attachment
				if len(event.Status.MediaAttachments) > 0 && event.Status.MediaAttachments[0].Type == "image" {
					fmt.Println("-> and it's got an image")

					imageData, err := downloadImage(event.Status.MediaAttachments[0].URL)
					if err != nil {
						fmt.Println("Failed to download image:", err)
						return
					}

					fmt.Println("Image downloaded")

					croppedImageData, err := autocropImage(imageData)
					if err != nil {
						fmt.Println("Failed to autocrop image:", err)
						return
					}

					fmt.Println("Image cropped")

					// Extract board dimensions from ALT data
					boardWidth, boardHeight, err := extractBoardDimensions(event.Status.MediaAttachments[0].Description)
					if err != nil {
						fmt.Println("Failed to extract board dimensions:", err)
						return
					}

					fmt.Println("Board dimensions extracted")

					imageGrid, err := imageToGridImages(croppedImageData, boardWidth, boardHeight)
					if err != nil {
						fmt.Println("Failed to convert image to grid images:", err)
						return
					}

					fmt.Println("Image converted to grid images")

					snakeSpaceGrid, err := convertImageGridToSnakeSpaceGrid(imageGrid)
					if err != nil {
						fmt.Println("Failed to convert image grid to SnakeSpace grid:", err)
						return
					}

					fmt.Println("Image grid converted to SnakeSpace grid")

					gameState, err := convertSnakeSpaceGridToGameState(snakeSpaceGrid)
					if err != nil {
						fmt.Println("Failed to convert SnakeSpace grid to game state:", err)
						return
					}

					fmt.Println("SnakeSpace grid converted to game state")

					// Get the best move
					bestMove := determineNextMove(gameState)

					fmt.Println("Best move determined")

					// Respond to the post with the chosen move
					makePost(client, event, snakeSpaceGrid, bestMove)

					fmt.Println("Post made")
				}
			}
		}
	}
}

func makePost(client *mastodon.Client, event *mastodon.UpdateEvent, snakeSpaceGrid [][]SnakeSpace, move string) {
	fmt.Println("Making mastodon post")

	// Implement post reply logic here
	gridStr := snakeSpaceGridAsString(snakeSpaceGrid)

	status := "I am watching snakebot slithering around the board. The most recent update I saw was "
	status += event.Status.URL
	status += "\n\n"
	status += "My eyes aren't always perfect, but I think I see this:\n\n"
	status += gridStr
	status += "\n\n"
	status += "A good move might be " + move + ". But I'm not super good at this yet, and there may be better moves."

	// Post a new message, referring to the original message by its URL
	_, err := client.PostStatus(context.Background(), &mastodon.Toot{
		Status: status,
	})
	if err != nil {
		log.Fatal(err)
	}
}
