package main

import "log"

// Called when you create your Battlesnake on play.battlesnake.com and controls
// your Battlesnake's appearance.
func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "reidemption",
		Color:      "#3f51b5",
		Head:       "default",
		Tail:       "default",
	}
}

func start(state GameState) {
	log.Println("GAME START")
}

func end(state GameState) {
	log.Printf("GAME OVER after %d turns\n\n", state.Turn)
}

func main() {
	RunServer()
}
