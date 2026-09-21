"use client";

import { useState } from "react";

type Pokemon = {
  id: number;
  name: string;
  types: string[];
  score: number;
};

type Draft = {
  choices: Pokemon[];
  team: Pokemon[];
  picks: number;
  complete: boolean;
  total_score: number;
};

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export default function DraftGame() {
  const [draft, setDraft] = useState<Draft | null>(null);
  const [message, setMessage] = useState("Start a draft to receive three Pokemon choices.");
  const [isLoading, setIsLoading] = useState(false);

  async function requestDraft(path: string, options?: RequestInit) {
    if (!apiURL) {
      throw new Error("NEXT_PUBLIC_API_URL is not configured.");
    }

    const response = await fetch(`${apiURL}${path}`, options);
    if (!response.ok) {
      throw new Error(await response.text());
    }

    return (await response.json()) as Draft;
  }

  async function startDraft() {
    setIsLoading(true);
    try {
      const nextDraft = await requestDraft("/api/draft", { method: "POST" });
      setDraft(nextDraft);
      setMessage("Choose one Pokemon. The backend will generate the next round.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not start draft.");
    } finally {
      setIsLoading(false);
    }
  }

  async function choosePokemon(pokemonID: number) {
    setIsLoading(true);
    try {
      const nextDraft = await requestDraft("/api/draft/picks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pokemon_id: pokemonID }),
      });
      setDraft(nextDraft);
      setMessage(nextDraft.complete ? "Draft complete!" : "Pokemon selected. Choose from the next round.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not save pick.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <section className="draft-game">
      <div className="draft-summary">
        <p>{message}</p>
        <p>Picked: {draft?.picks ?? 0}/6</p>
        <p>Total score: {draft?.total_score ?? 0}</p>
      </div>

      <button onClick={startDraft} disabled={isLoading}>
        {isLoading ? "Loading..." : "Start new draft"}
      </button>

      {draft && !draft.complete && (
        <>
          <h2>Choose one</h2>
          <div className="pokemon-grid">
            {draft.choices.map((pokemon) => (
              <article className="pokemon-card" key={pokemon.id}>
                <h3>{pokemon.name}</h3>
                <p>Types: {pokemon.types.join(", ")}</p>
                <p>Base-stat score: {pokemon.score}</p>
                <button disabled={isLoading} onClick={() => choosePokemon(pokemon.id)}>
                  Choose {pokemon.name}
                </button>
              </article>
            ))}
          </div>
        </>
      )}

      {draft && (
        <>
          <h2>Your team</h2>
          {draft.team.length === 0 ? (
            <p>No Pokemon selected yet.</p>
          ) : (
            <ol>
              {draft.team.map((pokemon) => (
                <li key={pokemon.id}>
                  {pokemon.name} — {pokemon.score}
                </li>
              ))}
            </ol>
          )}
        </>
      )}
    </section>
  );
}
