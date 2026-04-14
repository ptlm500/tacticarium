import { useParams, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import {
  type RestGameEvent,
  buildPlayerStats,
  getEndReason,
  getRoundsPlayed,
} from "../components/game/vpUtils";
import { VPBreakdownTable } from "../components/game/VPBreakdownTable";
import { VPProgressionChart } from "../components/game/VPProgressionChart";
import { EventTimeline } from "../components/game/EventTimeline";
import { normalizeRestEvent } from "../components/game/eventFormatting";
import { useGame, useGameEvents } from "../hooks/queries/useGamesQueries";

export function GameDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const { data: gameState, isPending: gameLoading } = useGame(id!);
  const { data: events, isPending: eventsLoading } = useGameEvents(id!);

  if (gameLoading || eventsLoading) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>Loading...</p>
      </div>
    );
  }

  if (!gameState || !events) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <div className="text-center space-y-4">
          <p className="text-red-400">Game not found</p>
          <button
            onClick={() => navigate("/history")}
            className="text-indigo-400 hover:text-indigo-300"
          >
            Back to History
          </button>
        </div>
      </div>
    );
  }

  const myPlayer = gameState.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState.players.find((p) => p?.userId !== user?.id) ?? null;

  if (!myPlayer) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p className="text-gray-400">You were not a player in this game.</p>
      </div>
    );
  }

  const typedEvents = events as RestGameEvent[];
  const endReason = getEndReason(typedEvents);
  const roundsPlayed = getRoundsPlayed(typedEvents);
  const rounds = Array.from({ length: roundsPlayed }, (_, i) => i + 1);

  const myStats = buildPlayerStats(typedEvents, myPlayer.playerNumber, myPlayer.vpPaint);
  const opponentStats = opponent
    ? buildPlayerStats(typedEvents, opponent.playerNumber, opponent.vpPaint)
    : null;

  const isAbandoned = gameState.status === "abandoned";
  const isWinner = gameState.winnerId === user?.id;
  let heading: string;
  if (isAbandoned) {
    heading = "Game Abandoned";
  } else if (!gameState.winnerId) {
    heading = "Draw";
  } else {
    heading = isWinner ? "Victory!" : "Defeat";
  }

  const headingColor = isAbandoned
    ? "text-gray-400"
    : !gameState.winnerId
      ? "text-yellow-400"
      : isWinner
        ? "text-green-400"
        : "text-red-400";

  const reasonLabel =
    endReason === "concede"
      ? "Ended by concession"
      : endReason === "abandoned"
        ? "Mutually abandoned"
        : endReason === "rounds_complete"
          ? `Completed after ${roundsPlayed} rounds`
          : null;

  const normalizedEvents = typedEvents.map(normalizeRestEvent);

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">Game Details</h1>
        <button onClick={() => navigate("/history")} className="text-gray-400 hover:text-white">
          Back
        </button>
      </header>

      <main className="max-w-2xl mx-auto p-6 space-y-6">
        {/* Result Header */}
        <div className="text-center">
          <h2 className={`text-2xl font-bold ${headingColor}`}>{heading}</h2>
          {reasonLabel && <p className="text-sm text-gray-400 mt-1">{reasonLabel}</p>}
          {gameState.missionName && (
            <p className="text-sm text-gray-500 mt-1">{gameState.missionName}</p>
          )}
          <p className="text-xs text-gray-600 mt-1">
            {myPlayer.factionName} vs {opponent?.factionName ?? "Unknown"}
          </p>
        </div>

        {/* VP Breakdown Table */}
        <section className="bg-gray-800 border border-gray-700 rounded-xl p-4">
          <h3 className="text-sm font-semibold text-gray-400 mb-3">VP Breakdown</h3>
          <VPBreakdownTable
            myStats={myStats}
            opponentStats={opponentStats}
            myUsername={myPlayer.username}
            opponentUsername={opponent?.username ?? null}
            rounds={rounds}
          />
        </section>

        {/* VP Progression Chart */}
        {rounds.length > 0 && (
          <section className="bg-gray-800 border border-gray-700 rounded-xl p-4">
            <h3 className="text-sm font-semibold text-gray-400 mb-3">VP Progression</h3>
            <VPProgressionChart
              myStats={myStats}
              opponentStats={opponentStats}
              myUsername={myPlayer.username}
              opponentUsername={opponent?.username ?? null}
              rounds={rounds}
            />
          </section>
        )}

        {/* Stats */}
        <div className="flex flex-wrap gap-6 text-sm text-gray-400">
          <div>
            Rounds played: <span className="text-white">{roundsPlayed}</span>
          </div>
          <div>
            {myPlayer.username} stratagems:{" "}
            <span className="text-white">{myStats.stratagemsUsed}</span>
          </div>
          {opponentStats && (
            <div>
              {opponent!.username} stratagems:{" "}
              <span className="text-white">{opponentStats.stratagemsUsed}</span>
            </div>
          )}
        </div>

        {/* Event Timeline */}
        <section className="bg-gray-800 border border-gray-700 rounded-xl p-4">
          <EventTimeline events={normalizedEvents} defaultFilter="highlights" />
        </section>
      </main>
    </div>
  );
}
