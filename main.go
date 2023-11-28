package main

import (
	"context"
	"fmt"
	_ "image/png"
	"log"
	"os"
	"time"

	"github.com/mattn/go-mastodon"
)

func main() {
	// Uncomment this to test the code without connecting to Mastodon
	// testMain()
	// return

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

	// Create a channel and goroutine to process votes on polls
	pollChannel := make(chan PollMessage)
	go processPolls(pollChannel, client)

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
				if event.Status.Poll != nil {
					fmt.Println("-> and it's got a poll")

					// Type of InReplyToID is interface{}, so we need to convert it to a string
					originalStatusIdStr := mastodon.ID(fmt.Sprintf("%v", event.Status.InReplyToID))

					// Send a message to the poll processing goroutine
					pollChannel <- PollMessage{
						MessageType: NewPoll,
						UpdateID:    originalStatusIdStr,
						PollID:      event.Status.Poll.ID,
						PollOptions: event.Status.Poll.Options,
					}

					// Run timer coroutine to wake up two minutes before the poll expires
					go func() {
						// Sleep until two minutes before the poll expires
						sleepTime := event.Status.Poll.ExpiresAt.Sub(event.Status.CreatedAt) - (2 * time.Minute)
						fmt.Println("⏰ Timer coroutine sleeping for", sleepTime)
						time.Sleep(sleepTime)

						fmt.Println("⏰ Timer coroutine woke up. Sending message to poll processing goroutine")

						// After waking up, send a message to the poll processing goroutine
						pollChannel <- PollMessage{
							MessageType: TimerCheck,
							UpdateID:    originalStatusIdStr,
							PollID:      event.Status.Poll.ID,
						}
					}()

				} else if len(event.Status.MediaAttachments) > 0 && event.Status.MediaAttachments[0].Type == "image" {
					// Process the image
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
					bestMove, mustTurn := determineNextMove(gameState)

					fmt.Println("Best move determined")

					// Respond to the post with the chosen move
					myUpdateId := makePost(client, event, snakeSpaceGrid, bestMove, mustTurn)

					// Tell the poll processing goroutine that we've made a post
					pollChannel <- PollMessage{
						MessageType: NewState,
						UpdateID:    event.Status.ID,
						MyVote:      bestMove,
						MustTurn:    mustTurn,
						MyUpdateId:  myUpdateId,
					}

					fmt.Println("Post made")
				}
			}
		}
	}
}

func makePost(client *mastodon.Client, event *mastodon.UpdateEvent, snakeSpaceGrid [][]SnakeSpace, move string, mustTurn bool) mastodon.ID {
	fmt.Println("Making mastodon post")

	// Implement post reply logic here
	gridStr := snakeSpaceGridAsString(snakeSpaceGrid)

	status := "I am watching snakebot slithering. The most recent update I saw was "
	status += event.Status.URL
	status += "\n\n"
	status += "This is what I see, in text form:\n\n"
	status += gridStr
	status += "\n\n"
	if mustTurn {
		status += "It looks like the snake is headed for disaster if it doesn't turn!\n\n"
	}
	status += "I'm not very smart, but I think the snake should move " + move + " next."

	// Post a new message, referring to the original message by its URL
	post, err := client.PostStatus(context.Background(), &mastodon.Toot{
		Status: status,
	})
	if err != nil {
		log.Fatal(err)
	}

	return post.ID
}
