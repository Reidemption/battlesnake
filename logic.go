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
	killBonus       = 500
	threatPenalty   = 1000
	proximityWeight = 10
	spaceWeight     = 20
	trapPenalty     = 100
	// Room beyond twice our length is indistinguishable in practice, and capping it
	// keeps an open board from swamping every other term.
	spaceHorizon = 2

	baseFoodWeight  = 4
	hungerThreshold = 50
	hungerScale     = 2
	// Must exceed proximityWeight, or losing the length race still loses to the urge to chase.
	contestedFoodWeight = 12
	// One piece of food must not flip a head-to-head we committed to on the previous turn.
	huntMargin = 2
	// Held under trapPenalty so starvation never argues us into a dead end.
	maxFoodWeight = 60
)

// Scores every legal neighbour of our head and takes the best, breaking ties at random.
func chooseMove(state GameState) BattlesnakeMoveResponse {
	occupied := occupiedSquares(state.Board.Snakes)
	target := huntTarget(state)

	bestScore := math.MinInt
	best := []string{}

	for _, dir := range directions {
		next := Coord{X: state.You.Head.X + dir.Delta.X, Y: state.You.Head.Y + dir.Delta.Y}
		if !inBounds(next, state.Board) || occupied[next] {
			continue
		}

		score := scoreMove(next, state, target, occupied)
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

func scoreMove(next Coord, state GameState, target *Battlesnake, occupied map[Coord]bool) int {
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
		// Squares adjacent to a head are contested: whoever is longer survives the trade.
		if manhattan(next, snake.Head) <= 1 {
			if snake.Length < state.You.Length {
				score += killBonus
			} else {
				score -= threatPenalty
			}
		}
	}

	if target != nil {
		score -= manhattan(next, target.Head) * proximityWeight
	}

	if food, ok := nearestFood(next, state.Board.Food); ok {
		behind := longestOpponent(state) >= state.You.Length
		score -= manhattan(next, food) * foodWeight(state.You.Health, behind)
	}

	return score
}

// Scales from mild growth pressure when fed and dominant to an overriding pull when
// starving or losing the length race.
func foodWeight(health int, behind bool) int {
	weight := baseFoodWeight
	if behind {
		weight += contestedFoodWeight
	}
	if health < hungerThreshold {
		weight += (hungerThreshold - health) * hungerScale
	}
	return minInt(weight, maxFoodWeight)
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

// Nearest snake we outlength by enough to survive the trade, or nil when chasing
// anything would cost us the length race instead.
func huntTarget(state GameState) *Battlesnake {
	if longestOpponent(state) >= state.You.Length {
		return nil
	}

	var target *Battlesnake
	closest := math.MaxInt

	for i := range state.Board.Snakes {
		snake := &state.Board.Snakes[i]
		if snake.ID == state.You.ID || snake.Length+huntMargin > state.You.Length {
			continue
		}
		if d := manhattan(state.You.Head, snake.Head); d < closest {
			closest = d
			target = snake
		}
	}

	return target
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

func longestOpponent(state GameState) int {
	longest := 0
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		longest = maxInt(longest, snake.Length)
	}
	return longest
}

func occupiedSquares(snakes []Battlesnake) map[Coord]bool {
	occupied := map[Coord]bool{}
	for _, snake := range snakes {
		// The tail vacates this turn unless the snake just ate, which we ignore here.
		for _, part := range snake.Body {
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
