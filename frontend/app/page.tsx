import HealthStatus from "./health-status";
import PokemonStatus from "./pokemon-status";

export default function Home() {
  return (
    <main>
      <h1>Poké Draft</h1>
      <p>Draft six Pokémon and build the highest-scoring team.</p>
      <HealthStatus />
      <PokemonStatus />
    </main>
  );
}
