import DraftGame from "./draft-game";
import HealthStatus from "./health-status";

export default function Home() {
  return (
    <main>
      <h1>Pokemon Draft</h1>
      <p>Draft six Pokemon and build the highest-scoring team.</p>
      <HealthStatus />
      <DraftGame />
    </main>
  );
}
