package main

import "testing"

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
