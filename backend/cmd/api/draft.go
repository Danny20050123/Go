package main

import (
	"context"
	"errors"
	"math/rand/v2"
)

const (
	teamSize        = 6
	choicesPerRound = 3
	maxPokemonID    = 1025
)

var (
	errDraftNotStarted = errors.New("draft has not started")
	errInvalidPick     = errors.New("pokemon is not one of the current choices")
	errDraftComplete   = errors.New("draft is complete")
)

type draftState struct {
	Choices [choicesPerRound]pokemon
	Team    [teamSize]pokemon
	Picks   int
}

type draftResponse struct {
	Choices    [choicesPerRound]pokemon `json:"choices"`
	Team       []pokemon                `json:"team"`
	Picks      int                      `json:"picks"`
	Complete   bool                     `json:"complete"`
	TotalScore int                      `json:"total_score"`
}

// generatePokemonIDs returns three different standard Pokédex IDs in the
// inclusive range 1 through 1025.
func generatePokemonIDs() [choicesPerRound]int {
	var ids [choicesPerRound]int

	for index := range ids {
		for {
			id := rand.IntN(maxPokemonID) + 1
			if !containsID(ids[:index], id) {
				ids[index] = id
				break
			}
		}
	}

	return ids
}

// generateChoices fetches data for three random Pokémon that have not already
// been selected for the team.
func (draft *draftState) generateChoices(ctx context.Context) error {
	ids := generatePokemonIDs()
	for index, id := range ids {
		for {
			if draft.hasPokemon(id) || containsID(choiceIDs(draft.Choices[:index]), id) {
				id = rand.IntN(maxPokemonID) + 1
				continue
			}

			pokemon, err := getPokemonByID(ctx, id)
			if err != nil {
				return err
			}

			draft.Choices[index] = pokemon
			break
		}
	}

	return nil
}

// choose stores a current choice in the next available team slot. The chosen
// Pokémon can never be picked again because later choices exclude the team.
func (draft *draftState) choose(ctx context.Context, pokemonID int) error {
	if draft.Picks == teamSize {
		return errDraftComplete
	}

	for _, choice := range draft.Choices {
		if choice.ID != pokemonID {
			continue
		}

		next := *draft
		next.Team[next.Picks] = choice
		next.Picks++
		if next.Picks == teamSize {
			next.Choices = [choicesPerRound]pokemon{}
			*draft = next
			return nil
		}

		if err := next.generateChoices(ctx); err != nil {
			return err
		}

		*draft = next
		return nil
	}

	return errInvalidPick
}

func (draft draftState) hasPokemon(pokemonID int) bool {
	for _, pokemon := range draft.Team[:draft.Picks] {
		if pokemon.ID == pokemonID {
			return true
		}
	}

	return false
}

func (draft draftState) response() draftResponse {
	team := make([]pokemon, draft.Picks)
	copy(team, draft.Team[:draft.Picks])

	response := draftResponse{
		Choices:  draft.Choices,
		Team:     team,
		Picks:    draft.Picks,
		Complete: draft.Picks == teamSize,
	}
	for _, pokemon := range team {
		response.TotalScore += pokemon.Score
	}

	return response
}

func containsID(ids []int, target int) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}

	return false
}

func choiceIDs(choices []pokemon) []int {
	ids := make([]int, len(choices))
	for index, choice := range choices {
		ids[index] = choice.ID
	}

	return ids
}
