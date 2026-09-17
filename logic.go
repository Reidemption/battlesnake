package main

import (
	"math"
	"math/rand"
)

// Fixed order keeps behaviour reproducible; ties are broken randomly instead.
var directions = []struct {
	Name  string
	Delta Coord
}{
	{"up", Coord{X: 0, Y: 1}},
	{"down", Coord{X: 0, Y: -1}},
	{"left", Coord{X: -1, Y: 0}},
	{"right", Coord{X: 1, Y: 0}},
}

const (
	// A contested square is only ever worth entering if we are longer, and this snake
	// never tries to be longer, so every contested square is refused.
	threatPenalty = 1000
	spaceWeight   = 20
	trapPenalty   = 100
	// Room beyond twice our length is indistinguishable in practice, and capping it
	// keeps an open board from swamping every other term.
	spaceHorizon = 2
	// Following our own tail is what produces the coil; kept well under the space terms
	// so it shapes the path instead of walking us into a pocket.
	tailWeight = 6
	// The ring we hold around the food we have claimed. Radius 2 is the smallest diamond
	// that never forces a step onto the food itself.
	orbitRadius = 2
	orbitWeight = 8
	// Health is a percentage, so this is the 25% mark where circling gives way to eating.
	eatThreshold     = 25
	eatWeight        = 40
	foodAvoidPenalty = 60
)

// Scores every legal neighbour of our head and takes the best, breaking ties at random.
func chooseMove(state GameState) BattlesnakeMoveResponse {
	occupied := occupiedSquares(state)
	// Picked from the head so every candidate is judged against the same food; picking it
	// per candidate would let the claim flip direction mid-orbit.
	target, hasTarget := nearestFood(state.You.Head, state.Board.Food)

	bestScore := math.MinInt
	best := []string{}

	for _, dir := range directions {
		next := Coord{X: state.You.Head.X + dir.Delta.X, Y: state.You.Head.Y + dir.Delta.Y}
		if !inBounds(next, state.Board) || occupied[next] {
			continue
		}

		score := scoreMove(next, state, target, hasTarget, occupied)
		if score > bestScore {
			bestScore = score
			best = []string{dir.Name}
		} else if score == bestScore {
			best = append(best, dir.Name)
		}
	}

	if len(best) == 0 {
		return BattlesnakeMoveResponse{Move: "up", Shout: "goodbye"}
	}
	return BattlesnakeMoveResponse{Move: best[rand.Intn(len(best))]}
}

func scoreMove(next Coord, state GameState, target Coord, hasTarget bool, occupied map[Coord]bool) int {
	score := 0

	reachable := reachableSpace(next, state.Board, occupied)
	score += minInt(reachable, state.You.Length*spaceHorizon) * spaceWeight
	if reachable < state.You.Length {
		score -= (state.You.Length - reachable) * trapPenalty
	}

	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		if manhattan(next, snake.Head) <= 1 {
			score -= threatPenalty
		}
	}

	if tail, ok := ownTail(state.You); ok {
		score -= manhattan(next, tail) * tailWeight
	}

	if hasTarget {
		distance := manhattan(next, target)
		if state.You.Health <= eatThreshold {
			score -= distance * eatWeight
		} else {
			score -= abs(distance-orbitRadius) * orbitWeight
			if isFood(next, state.Board.Food) {
				// Stepping on food is not optional, so an unwanted segment can only be
				// declined by refusing the square.
				score -= foodAvoidPenalty
			}
		}
	}

	return score
}

func isFood(c Coord, food []Coord) bool {
	for _, f := range food {
		if f == c {
			return true
		}
	}
	return false
}

func ownTail(you Battlesnake) (Coord, bool) {
	if len(you.Body) < 2 {
		return Coord{}, false
	}
	return you.Body[len(you.Body)-1], true
}

func nearestFood(from Coord, food []Coord) (Coord, bool) {
	closest := math.MaxInt
	var found Coord

	for _, f := range food {
		if d := manhattan(from, f); d < closest {
			closest = d
			found = f
		}
	}

	return found, closest < math.MaxInt
}

// Squares connected to start through free ground, start included.
func reachableSpace(start Coord, board Board, occupied map[Coord]bool) int {
	if !inBounds(start, board) || occupied[start] {
		return 0
	}

	seen := map[Coord]bool{start: true}
	queue := []Coord{start}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, dir := range directions {
			next := Coord{X: cur.X + dir.Delta.X, Y: cur.Y + dir.Delta.Y}
			if !inBounds(next, board) || occupied[next] || seen[next] {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}

	return len(seen)
}

func occupiedSquares(state GameState) map[Coord]bool {
	occupied := map[Coord]bool{}

	for _, snake := range state.Board.Snakes {
		body := snake.Body
		// Our own tail moves off its square as we advance, which is the whole reason
		// following it is legal. Health at full means we just ate and it stays put.
		if snake.ID == state.You.ID && len(body) > 1 && snake.Health < 100 {
			body = body[:len(body)-1]
		}
		for _, part := range body {
			occupied[part] = true
		}
	}

	return occupied
}

func inBounds(c Coord, board Board) bool {
	return c.X >= 0 && c.X < board.Width && c.Y >= 0 && c.Y < board.Height
}

func manhattan(a, b Coord) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// The min and max builtins arrived in Go 1.21; replit.nix pins Go 1.17.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
