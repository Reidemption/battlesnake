package main

import "testing"

func TestMoveHuntsShorterSnake(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 5, Y: 4}, {X: 5, Y: 3}, {X: 5, Y: 2}, {X: 5, Y: 1}},
		Length: 5,
	}
	prey := Battlesnake{
		ID:     "them",
		Head:   Coord{X: 8, Y: 5},
		Body:   []Coord{{X: 8, Y: 5}, {X: 8, Y: 4}},
		Length: 2,
	}
	state := GameState{
		Board: Board{Width: 11, Height: 11, Snakes: []Battlesnake{me, prey}},
		You:   me,
	}

	if got := move(state).Move; got != "right" {
		t.Errorf("expected to close on prey with right, got %q", got)
	}
}

func TestMoveAvoidsLongerSnakeHead(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 5, Y: 4}},
		Length: 2,
	}
	bully := Battlesnake{
		ID:     "them",
		Head:   Coord{X: 7, Y: 5},
		Body:   []Coord{{X: 7, Y: 5}, {X: 7, Y: 4}, {X: 7, Y: 3}},
		Length: 3,
	}
	state := GameState{
		Board: Board{Width: 11, Height: 11, Snakes: []Battlesnake{me, bully}},
		You:   me,
	}

	for i := 0; i < 20; i++ {
		if got := move(state).Move; got == "right" {
			t.Fatalf("stepped into a longer snake's reach on iteration %d", i)
		}
	}
}

// Head at (5,5) with an opponent walling off a two-square pocket at (4,5)-(4,6).
func pocketState() GameState {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 6, Y: 5}, {X: 7, Y: 5}},
		Length: 3,
	}
	wall := Battlesnake{
		ID:   "them",
		Head: Coord{X: 5, Y: 6},
		Body: []Coord{
			{X: 5, Y: 6}, {X: 5, Y: 7}, {X: 4, Y: 7}, {X: 3, Y: 7},
			{X: 3, Y: 6}, {X: 3, Y: 5}, {X: 3, Y: 4}, {X: 4, Y: 4},
		},
		Length: 8,
	}
	return GameState{
		Board: Board{Width: 11, Height: 11, Snakes: []Battlesnake{me, wall}},
		You:   me,
	}
}

func TestReachableSpaceCountsSealedPocket(t *testing.T) {
	state := pocketState()
	occupied := occupiedSquares(state.Board.Snakes)

	if got := reachableSpace(Coord{X: 4, Y: 5}, state.Board, occupied); got != 2 {
		t.Errorf("expected a sealed pocket of 2, got %d", got)
	}
}

func TestMoveRejectsPocketForOpenGround(t *testing.T) {
	if got := move(pocketState()).Move; got != "down" {
		t.Errorf("expected open ground via down, got %q", got)
	}
}

func TestFoodWeightScalesWithHunger(t *testing.T) {
	cases := []struct {
		health int
		behind bool
		want   int
	}{
		{100, false, 4},
		{50, false, 4},
		{40, false, 24},
		{10, false, 60},
		{0, false, 60},
		{100, true, 16},
		{40, true, 36},
		{10, true, 60},
	}
	for _, c := range cases {
		if got := foodWeight(c.health, c.behind); got != c.want {
			t.Errorf("foodWeight(%d, %v) = %d, want %d", c.health, c.behind, got, c.want)
		}
	}
}

// The scenario that motivated huntMargin: a one-segment lead is not a lead, because the
// opponent eating once turns our chase into a losing head-to-head.
func TestMoveEatsInsteadOfChasingNarrowLead(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 5, Y: 4}, {X: 5, Y: 3}},
		Length: 3,
		Health: 100,
	}
	rival := Battlesnake{
		ID:     "them",
		Head:   Coord{X: 8, Y: 5},
		Body:   []Coord{{X: 8, Y: 5}, {X: 8, Y: 4}},
		Length: 2,
	}
	state := GameState{
		Board: Board{
			Width: 11, Height: 11,
			Food:   []Coord{{X: 2, Y: 5}},
			Snakes: []Battlesnake{me, rival},
		},
		You: me,
	}

	if huntTarget(state) != nil {
		t.Error("expected no hunt target on a one-segment lead")
	}
	if got := move(state).Move; got != "left" {
		t.Errorf("expected to break off toward food with left, got %q", got)
	}
}

func TestHuntSuppressedWhenOutlengthed(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 5, Y: 4}, {X: 5, Y: 3}},
		Length: 3,
	}
	runt := Battlesnake{ID: "runt", Head: Coord{X: 8, Y: 5}, Length: 1}
	giant := Battlesnake{ID: "giant", Head: Coord{X: 1, Y: 1}, Length: 9}
	state := GameState{
		Board: Board{Width: 11, Height: 11, Snakes: []Battlesnake{me, runt, giant}},
		You:   me,
	}

	if huntTarget(state) != nil {
		t.Error("expected growth to take priority while a longer snake is on the board")
	}
}

func TestMoveSeeksFoodWhenStarving(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 5, Y: 4}, {X: 5, Y: 3}},
		Length: 3,
		Health: 10,
	}
	state := GameState{
		Board: Board{
			Width: 11, Height: 11,
			Food:   []Coord{{X: 8, Y: 5}},
			Snakes: []Battlesnake{me},
		},
		You: me,
	}

	if got := move(state).Move; got != "right" {
		t.Errorf("expected to close on food with right, got %q", got)
	}
}

func TestMoveFavorsHuntOverFoodWhenFed(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 5, Y: 5},
		Body:   []Coord{{X: 5, Y: 5}, {X: 5, Y: 4}, {X: 5, Y: 3}, {X: 5, Y: 2}, {X: 5, Y: 1}},
		Length: 5,
		Health: 100,
	}
	prey := Battlesnake{
		ID:     "them",
		Head:   Coord{X: 8, Y: 5},
		Body:   []Coord{{X: 8, Y: 5}, {X: 8, Y: 4}},
		Length: 2,
	}
	state := GameState{
		Board: Board{
			Width: 11, Height: 11,
			Food:   []Coord{{X: 2, Y: 5}},
			Snakes: []Battlesnake{me, prey},
		},
		You: me,
	}

	if got := move(state).Move; got != "right" {
		t.Errorf("expected to keep chasing prey with right, got %q", got)
	}
}

func TestMoveBoxedIn(t *testing.T) {
	me := Battlesnake{
		ID:     "me",
		Head:   Coord{X: 0, Y: 0},
		Body:   []Coord{{X: 0, Y: 0}},
		Length: 1,
	}
	state := GameState{
		Board: Board{Width: 1, Height: 1, Snakes: []Battlesnake{me}},
		You:   me,
	}

	got := move(state)
	if got.Move != "up" || got.Shout != "goodbye" {
		t.Errorf("expected the default surrender move, got %+v", got)
	}
}
