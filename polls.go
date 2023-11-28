package main

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/mattn/go-mastodon"
)

type PollMessageType int64

const (
	UndefPollMsg PollMessageType = iota
	NewState
	NewPoll
	TimerCheck
)

// Structure for messages to send to the goroutine which votes on polls
type PollMessage struct {
	MessageType PollMessageType
	UpdateID    mastodon.ID
	MyVote      string
	MustTurn    bool
	PollID      mastodon.ID
	MyUpdateId  mastodon.ID
	PollOptions []mastodon.PollOption
}

type PollProcessState int64

const (
	UndefPollState PollProcessState = iota
	WaitingForState
	WaitingForPoll
	WaitingForTimer
)

// This runs as a worker goroutine which processes poll votes
func processPolls(pollChannel chan PollMessage, client *mastodon.Client) {
	fmt.Println("💪 Starting poll vote processing goroutine")

	currentState := UndefPollState
	var currentUpdateID mastodon.ID
	myVote := ""
	mustTurn := false
	var currentPollID mastodon.ID
	var myUpdateId mastodon.ID
	optionLookup := make(map[string]int)

	regexMove, err := regexp.Compile(`[Mm]ove (\w+)`)
	if err != nil {
		log.Fatal("💪 Error compiling regexp:", err)
	}

	for {
		if currentState == UndefPollState {
			fmt.Println("💪 Syncing to initial poll state")

			// Reset the state
			currentUpdateID = ""
			currentPollID = ""
			myUpdateId = ""
			myVote = ""
			mustTurn = false
			for k := range optionLookup {
				delete(optionLookup, k)
			}
			currentState = WaitingForState
		}

		fmt.Println("💪 Waiting for message in processPolls goroutine. Current state is", currentState)

		message := <-pollChannel
		fmt.Println("💪 Received poll message:", message)

		switch message.MessageType {
		case NewState:
			if currentState != WaitingForState {
				fmt.Println("💪 Received new state message while not waiting for start. Resetting state.")
				currentState = UndefPollState
				continue
			}
			currentUpdateID = message.UpdateID
			myVote = message.MyVote
			mustTurn = message.MustTurn
			currentState = WaitingForPoll
			myUpdateId = message.MyUpdateId
		case NewPoll:
			if currentState != WaitingForPoll {
				fmt.Println("💪 Received new poll message while not waiting for poll. Resetting state.")
				currentState = UndefPollState
				continue
			}
			if message.UpdateID != currentUpdateID {
				fmt.Println("💪 Received new poll message with different update ID. Resetting state.")
				currentState = UndefPollState
				continue
			}
			currentPollID = message.PollID
			for i, option := range message.PollOptions {
				matchMove := regexMove.FindStringSubmatch(option.Title)
				if len(matchMove) != 2 {
					fmt.Println("💪 Error parsing move from poll option title:", option.Title)
					continue
				}
				direction := matchMove[1]
				optionLookup[direction] = i
			}
			currentState = WaitingForTimer
		case TimerCheck:
			if currentState != WaitingForTimer {
				fmt.Println("💪 Received timer check message while not waiting for timer. Resetting state.")
				currentState = UndefPollState
				continue
			}
			if message.PollID != currentPollID {
				fmt.Println("💪 Received timer check message with different poll ID. Resetting state.")
				currentState = UndefPollState
				continue
			}

			poll, err := client.GetPoll(context.Background(), mastodon.ID(currentPollID))
			if err != nil {
				fmt.Println("💪 Error getting poll:", err)
				currentState = UndefPollState
				continue
			}

			fmt.Println("💪 Got poll and counted votes:", poll.VotesCount)

			// If there is already at least one vote, don't vote
			if poll.VotesCount > 0 {
				fmt.Println("💪 There are already votes. Not voting.")
				currentState = UndefPollState
				continue
			}

			// Vote for the option that matches myVote
			if mustTurn && myVote != "" {
				// Post a message to mastodon saying that we're voting
				msg := "Nobody has voted and I am worried that the snake is doomed if nothing is done! "
				msg += "I usually don't vote, but this time, I'm voting to move " + myVote + "."
				_, err := client.PostStatus(context.Background(), &mastodon.Toot{
					Status:      msg,
					InReplyToID: myUpdateId,
				})
				if err != nil {
					fmt.Println("💪 Error posting status:", err)
				}

				fmt.Println("💪 Posted message to mastodon about my vote")

				// Vote
				vote := optionLookup[myVote]
				fmt.Println("💪 Voting for option", vote)

				_, err = client.PollVote(context.Background(), currentPollID, vote)
				if err != nil {
					fmt.Println("💪 Error voting:", err)
				}
			}

			// Reset the state
			currentState = UndefPollState
		}
	}
}
