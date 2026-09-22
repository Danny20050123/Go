package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratePokemonIDsReturnsThreeUniqueIDsInRange(t *testing.T) {
	for range 20 {
		ids := generatePokemonIDs()
		seen := make(map[int]bool)

		for _, id := range ids {
			if id < 1 || id > maxPokemonID {
				t.Fatalf("ID %d is outside the valid range", id)
			}
			if seen[id] {
				t.Fatalf("ID %d was generated more than once", id)
			}
			seen[id] = true
		}
	}
}

func TestStartDraftRequiresPlayerID(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/draft", strings.NewReader(`{}`))
	response := httptest.NewRecorder()

	startDraftHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestDraftResponseIncludesPlayerID(t *testing.T) {
	state := draftState{PlayerID: "player-42"}

	if response := state.response(); response.PlayerID != "player-42" {
		t.Fatalf("player ID = %q, want player-42", response.PlayerID)
	}
}

func TestPokemonCacheKeyNormalizesPokemonName(t *testing.T) {
	key := pokemonCacheKey("  Pikachu  ")
	if key != "pokemon:v1:pikachu" {
		t.Fatalf("cache key = %q, want pokemon:v1:pikachu", key)
	}
}
