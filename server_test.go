package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const soloBody = `{"you":{"id":"me","head":{"x":5,"y":5},"body":[{"x":5,"y":5}],"length":1,"health":100},` +
	`"board":{"width":11,"height":11,"snakes":[{"id":"me","head":{"x":5,"y":5},"body":[{"x":5,"y":5}],"length":1}]}}`

// Routing broke once when the go.mod directive changed under method-based patterns and
// every unit test still passed, so the wiring is asserted over real requests.
func TestRoutes(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	cases := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/", http.StatusOK},
		{http.MethodPost, "/start", http.StatusOK},
		{http.MethodPost, "/move", http.StatusOK},
		{http.MethodPost, "/end", http.StatusOK},
		{http.MethodGet, "/move", http.StatusMethodNotAllowed},
	}

	for _, c := range cases {
		req, err := http.NewRequest(c.method, srv.URL+c.path, strings.NewReader(soloBody))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != c.want {
			t.Errorf("%s %s = %d, want %d", c.method, c.path, resp.StatusCode, c.want)
		}
	}
}

func TestIndexReportsAPIVersion(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got BattlesnakeInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.APIVersion != "1" {
		t.Errorf("apiversion = %q, want \"1\"", got.APIVersion)
	}
}

func TestMoveEndpointReturnsLegalDirection(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/move", "application/json", strings.NewReader(soloBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got BattlesnakeMoveResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	switch got.Move {
	case "up", "down", "left", "right":
	default:
		t.Errorf("move = %q, want a direction", got.Move)
	}
}
