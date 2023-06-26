# snakebot_admirer
Code for a bot that keeps an eye on snakebot on mastodon.

Check me out on <a rel="me" href="https://botsin.space/@snakebot_admirer">Mastodon</a>

## Overview of operation

The bot listens for updates and polls from the snakebot user. That bot is followed by this bot so that it appears on this user's timeline.

When there is a game state update, it downloads the png image, decodes it, crops it (the image has some empty alpha channel pixels on the sides), and chops it into individual images of the grid size (which is extracted from the ALT text of the image). It then looks for the snake color, then looks at the edges of the grid image to find adjacencies, i.e. where the snake moved to and from on its way to this spot. It uses the black of the snake eyes to find the head of the snake.

Note: The author of the snake bot tells me that there is a hidden encoding of the game state in the text of the update, so all this image processing is unnecessary. Oops. Well, maybe some time I will replace the image processing code and use the embedded state data. It would be much more efficient and robust.

After the image processing, there is a n x m grid (8x5 is the only size I've seen snakebot generate) of data including empty spots, a food spot, and snake spots with adjaceny data. For analysis, it's easier to work with positional data, e.g. the food is at (X, Y) and the snake is defined by an array of (X, Y) points from head to tail. The food is found with a simple search of the grid. Building the snake array involves a slightly nifty algorithm that starts by finding the head and then following the adjacency data, which is encoded as a bitmask. At each point in the snake, we mask out the adkacency from which we came and move to the remaining adjacency. This is repeated until we hit the tail, which only has 1 adjacency.

After the game state is in this form, the AI is run to determine the prefered next move. See "AI" below for details.

The bot also tries to keep the snake alive when nobody is voting. I want to leave voting up to people for the most part. But sometimes nobody is paying attention to the bot (seems to happen mostly overnight in the US) and it's sad to watch the bot die when it's been working hard to stay alive for days. So if nobody has voted a couple minutes before the poll expires, it will vote based on what the AI algorithm thinks is best.

Voting is accomplished with 2 goroutines: one worker and one timer. The timer routine is created when the poll is posted, and it simply sleeps until 2 minutes before the poll expires, then sends a message to the worker to do its thing. The other messages the worker listens for come from the main thread, first when the game state update is posted by the snake bot (to get the AI's best move), then when the poll update is posted by the snake bot (to get the poll ID). When the timer message comes in, the worker uses the poll ID to get the poll data to see if anyone has voted. If nobody has voted, it posts a message saying it's voting and then it votes. I am considering adding functionality to make it vote only if the snake is going to crash if nobody votes.

## AI

The AI is currently quite simple, although I hope to improve it.

It simulates one move in each direction. If it crashes, the move is considered bad and given a negative score. If it doesn't crash, it gets a score where shorter manhattan distance to the food is preferred. The one other safety measure is that it considers moves where it can't reach its own tail to be bad. This should prevent it closing itself off in a space smaller than itself.

This is a reasonably safe, simple strategy, but very suboptimal in "late game" situations in terms of efficiency. In particular, it will go directly towards the food in situations where it will be forced to turn away from the food on the following turn. This will result in the snake having to follow its own tail for many turns until it can safely dart again at the food without losing sight of its tail.

I'm trying to figure out a better algorithm without using the internet to tell me the solution.

I may be overthinking it, but it seems like an optimal solution would be something akin to a chess search algorithm, where every possible path to the food is simulated. Upon reaching the food, the "opponent" is simulated (random food placement). This would happen recursively until a deep search reveals the best immediate move. I'm guessing this would lend itself to the clearly beneficial "coiling" behavior that people choose in late game scenarios. It seems like it may be overkill brute force, but the search space naturally shrinks on each recursion in a way that reminds me of brute force sudoku solvers, which are not computationally that expensive. Fun stuff to ponder. (This is probably a google interview question that some dork expects a full optimal solution in perfect idiomatic Python in 30 minutes to weed out dummies like myself.)

## Running

Here's a bash script which sets up the mastodon credentials and then runs it.

```bash
#!/usr/bin/env bash
export MASTODON_SERVER="https://<hostname of mastodon instance>"
export CLIENT_KEY="<mastodon app client key>"
export CLIENT_SECRET="<mastodon app client secret>"
export ACCESS_TOKEN="<mastodon app access token>"

# Run the application
go run .
```
