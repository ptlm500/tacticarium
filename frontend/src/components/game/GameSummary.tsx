import { useEffect, useState } from "react";
import { GameState, PlayerState } from "../../types/game";
import { gamesApi } from "../../api/games";
import { type RestGameEvent, buildPlayerStats, getEndReason, getRoundsPlayed } from "./vpUtils";
import { VPBreakdownTable } from "./VPBreakdownTable";

interface Props {
  gameState: GameState;
  myPlayer: PlayerState;
  opponent: PlayerState | null;
  currentUserId: string;
}

export function GameSummary({ gameState, myPlayer, opponent, currentUserId }: Props) {
  const [events, setEvents] = useState<RestGameEvent[] | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    gamesApi
      .getEvents(gameState.gameId)
      .then((data) => setEvents(data as RestGameEvent[]))
      .catch(() => setEvents([]))
      .finally(() => setLoading(false));
  }, [gameState.gameId]);

  const isWinner = gameState.winnerId === currentUserId;
  const isAbandoned = gameState.status === "abandoned";

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

  if (loading || !events) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>Loading summary...</p>
      </div>
    );
  }

  const endReason = getEndReason(events);
  const roundsPlayed = getRoundsPlayed(events);
  const rounds = Array.from({ length: roundsPlayed }, (_, i) => i + 1);

  const myStats = buildPlayerStats(events, myPlayer.playerNumber, myPlayer.vpPaint);
  const opponentStats = opponent
    ? buildPlayerStats(events, opponent.playerNumber, opponent.vpPaint)
    : null;

  const reasonLabel =
    endReason === "concede"
      ? "Ended by concession"
      : endReason === "abandoned"
        ? "Mutually abandoned"
        : endReason === "rounds_complete"
          ? `Completed after ${roundsPlayed} rounds`
          : null;

  return (
    <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center p-4">
      <div className="bg-gray-800 border border-gray-700 rounded-xl max-w-2xl w-full">
        {/* Header */}
        <div className="px-6 py-5 border-b border-gray-700 text-center">
          <h1 className={`text-3xl font-bold ${headingColor}`}>{heading}</h1>
          {reasonLabel && <p className="text-sm text-gray-400 mt-1">{reasonLabel}</p>}
        </div>

        {/* VP Breakdown Table */}
        <div className="px-6 py-4">
          <VPBreakdownTable
            myStats={myStats}
            opponentStats={opponentStats}
            myUsername={myPlayer.username}
            opponentUsername={opponent?.username ?? null}
            rounds={rounds}
          />
        </div>

        {/* Stats */}
        <div className="px-6 py-4 border-t border-gray-700 flex flex-wrap gap-6 text-sm text-gray-400">
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

        {/* Back to lobby */}
        <div className="px-6 py-4 border-t border-gray-700 text-center">
          <a
            href="/"
            className="inline-block bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2.5 px-6 rounded-lg transition-colors"
          >
            Back to Lobby
          </a>
        </div>
      </div>
    </div>
  );
}
