package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	frontendOrigin = "http://localhost:3000"
	pokeAPIURL     = "https://pokeapi.co/api/v2/pokemon/"
)

var pokeAPIClient = &http.Client{Timeout: 5 * time.Second}

var errPokemonNotFound = errors.New("pokemon not found")

var currentDraft struct {
	sync.Mutex
	state   draftState
	started bool
}

type pokemonResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Moves []struct {
		Move struct {
			Name string `json:"name"`
		} `json:"move"`
	} `json:"moves"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
}

type baseStats struct {
	HP             int `json:"hp"`
	Attack         int `json:"attack"`
	Defense        int `json:"defense"`
	SpecialAttack  int `json:"special_attack"`
	SpecialDefense int `json:"special_defense"`
	Speed          int `json:"speed"`
}

type pokemon struct {
	ID    int       `json:"id"`
	Name  string    `json:"name"`
	Types []string  `json:"types"`
	Moves []string  `json:"moves"`
	Stats baseStats `json:"stats"`
	Score int       `json:"score"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("GET /api/pokemon/{name}", pokemonHandler)
	mux.HandleFunc("POST /api/draft", startDraftHandler)
	mux.HandleFunc("GET /api/draft", getDraftHandler)
	mux.HandleFunc("POST /api/draft/picks", pickPokemonHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: cors(mux),
	}

	log.Printf("API listening on http://localhost%s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func pokemonHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(strings.TrimSpace(r.PathValue("name")))
	if name == "" {
		http.Error(w, "pokemon name is required", http.StatusBadRequest)
		return
	}

	pokemon, err := getPokemon(r.Context(), name)
	if err != nil {
		if errors.Is(err, errPokemonNotFound) {
			http.Error(w, "pokemon not found", http.StatusNotFound)
			return
		}

		log.Printf("get pokemon %q: %v", name, err)
		http.Error(w, "pokemon service unavailable", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, pokemon)
}

func startDraftHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID string `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	playerID := strings.TrimSpace(request.PlayerID)
	if playerID == "" {
		http.Error(w, errPlayerIDRequired.Error(), http.StatusBadRequest)
		return
	}

	currentDraft.Lock()
	defer currentDraft.Unlock()

	state := draftState{PlayerID: playerID}
	if err := state.generateChoices(r.Context()); err != nil {
		log.Printf("start draft: %v", err)
		http.Error(w, "pokemon service unavailable", http.StatusBadGateway)
		return
	}

	currentDraft.state = state
	currentDraft.started = true
	writeJSON(w, http.StatusCreated, currentDraft.state.response())
}

func getDraftHandler(w http.ResponseWriter, r *http.Request) {
	currentDraft.Lock()
	defer currentDraft.Unlock()

	if !currentDraft.started {
		http.Error(w, errDraftNotStarted.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, currentDraft.state.response())
}

func pickPokemonHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PokemonID int `json:"pokemon_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	currentDraft.Lock()
	defer currentDraft.Unlock()

	if !currentDraft.started {
		http.Error(w, errDraftNotStarted.Error(), http.StatusNotFound)
		return
	}

	if err := currentDraft.state.choose(r.Context(), request.PokemonID); err != nil {
		switch {
		case errors.Is(err, errInvalidPick):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, errDraftComplete):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			log.Printf("pick pokemon: %v", err)
			http.Error(w, "pokemon service unavailable", http.StatusBadGateway)
		}
		return
	}

	writeJSON(w, http.StatusOK, currentDraft.state.response())
}

func getPokemon(ctx context.Context, name string) (pokemon, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, pokeAPIURL+url.PathEscape(name), nil)
	if err != nil {
		return pokemon{}, fmt.Errorf("create request: %w", err)
	}

	response, err := pokeAPIClient.Do(request)
	if err != nil {
		return pokemon{}, fmt.Errorf("request pokeapi: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return pokemon{}, errPokemonNotFound
	}
	if response.StatusCode != http.StatusOK {
		return pokemon{}, fmt.Errorf("pokeapi returned status %d", response.StatusCode)
	}

	var source pokemonResponse
	if err := json.NewDecoder(response.Body).Decode(&source); err != nil {
		return pokemon{}, fmt.Errorf("decode pokeapi response: %w", err)
	}

	result := pokemon{ID: source.ID, Name: source.Name}
	for _, pokemonType := range source.Types {
		result.Types = append(result.Types, pokemonType.Type.Name)
	}
	for _, move := range source.Moves {
		result.Moves = append(result.Moves, move.Move.Name)
	}
	for _, stat := range source.Stats {
		switch stat.Stat.Name {
		case "hp":
			result.Stats.HP = stat.BaseStat
		case "attack":
			result.Stats.Attack = stat.BaseStat
		case "defense":
			result.Stats.Defense = stat.BaseStat
		case "special-attack":
			result.Stats.SpecialAttack = stat.BaseStat
		case "special-defense":
			result.Stats.SpecialDefense = stat.BaseStat
		case "speed":
			result.Stats.Speed = stat.BaseStat
		}
	}

	result.Score = result.Stats.HP + result.Stats.Attack + result.Stats.Defense +
		result.Stats.SpecialAttack + result.Stats.SpecialDefense + result.Stats.Speed

	return result, nil
}

func getPokemonByID(ctx context.Context, id int) (pokemon, error) {
	return getPokemon(ctx, strconv.Itoa(id))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == frontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
