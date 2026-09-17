"use client";

import { useEffect, useState } from "react";

type PokemonResponse = {
  id: number;
  name: string;
  score: number;
  types: string[];
  moves: string[];
  stats: {
    hp: number;
    attack: number;
    defense: number;
    special_attack: number;
    special_defense: number;
    speed: number;
  };
};

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export default function PokemonStatus() {
  const [message, setMessage] = useState("Loading Pikachu from the Go API...");
  const [details, setDetails] = useState("");

  useEffect(() => {
    if (!apiURL) {
      setMessage("NEXT_PUBLIC_API_URL is not configured.");
      return;
    }

    async function getPokemon() {
      try {
        const response = await fetch(`${apiURL}/api/pokemon/pikachu`);
        if (!response.ok) {
          throw new Error(`backend returned ${response.status}`);
        }

        const pokemon: PokemonResponse = await response.json();
        setMessage(
          `${pokemon.name} (#${pokemon.id}) has a backend-calculated base-stat score of ${pokemon.score}.`,
        );
        setDetails(
          `Types: ${pokemon.types.join(", ")}. Learnable moves: ${pokemon.moves.join(", ")}.`,
        );
      } catch (error) {
        const detail = error instanceof Error ? error.message : "unknown error";
        setMessage(`Pokemon unavailable: ${detail}`);
      }
    }

    void getPokemon();
  }, []);

  return (
    <section>
      <p>{message}</p>
      <p>{details}</p>
    </section>
  );
}
