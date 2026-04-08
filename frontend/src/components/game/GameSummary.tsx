import { useEffect, useState } from "react";
import { GameState, PlayerState } from "../../types/game";
import { gamesApi } from "../../api/games";

/** Event shape returned by the REST /api/games/:id/events endpoint. */
interface RestGameEvent {
  id: number;
  playerNumber: number | null;
  eventType: string;
  eventData: Record<string, unknown> | null;
  round: number | null;
  phase: string | null;
  createdAt: string;
}

interface Props {
  gameState: GameState;
  myPlayer: PlayerState;
  opponent: PlayerState | null;
  currentUserId: string;
}

interface RoundVP {
  primary: number;
  secondary: number;
  gambit: number;
}

interface PlayerSummaryStats {
  totalVP: number;
  vpByRound: Record<number, RoundVP>;
  stratagemsUsed: number;
  paint: number;
}

function buildPlayerStats(
  events: RestGameEvent[],
  playerNumber: number,
  paint: number,
): PlayerSummaryStats {
  const vpByRound: Record<number, RoundVP> = {};
  let stratagemsUsed = 0;

  for (const e of events) {
    if (e.playerNumber !== playerNumber) continue;
    const round = e.round ?? 0;

    if (
      e.eventType === "vp_primary_score" ||
      e.eventType === "vp_secondary_score" ||
      e.eventType === "vp_gambit_score" ||
      e.eventType === "secondary_achieved" ||
      e.eventType === "challenger_scored"
    ) {
      if (!vpByRound[round]) vpByRound[round] = { primary: 0, secondary: 0, gambit: 0 };

      const delta = (e.eventData?.delta as number) ?? 0;
      const vpScored = (e.eventData?.vpScored as number) ?? 0;
      const amount = delta || vpScored;

      if (e.eventType === "vp_primary_score") {
        vpByRound[round].primary += amount;
      } else if (e.eventType === "vp_secondary_score" || e.eventType === "secondary_achieved") {
        vpByRound[round].secondary += amount;
      } else if (e.eventType === "vp_gambit_score" || e.eventType === "challenger_scored") {
        vpByRound[round].gambit += amount;
      }
    }

    if (e.eventType === "stratagem_used") {
      stratagemsUsed++;
    }
  }

  let totalVP = paint;
  for (const rv of Object.values(vpByRound)) {
    totalVP += rv.primary + rv.secondary + rv.gambit;
  }

  return { totalVP, vpByRound, stratagemsUsed, paint };
}

function getEndReason(events: RestGameEvent[]): string | null {
  const endEvent = events.find((e) => e.eventType === "game_end");
  return (endEvent?.eventData?.reason as string) ?? null;
}

function getRoundsPlayed(events: RestGameEvent[]): number {
  let max = 0;
  for (const e of events) {
    if (e.round != null && e.round > max) max = e.round;
  }
  return max;
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
        <div className="px-6 py-4 overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left py-2 text-gray-400 font-medium">Round</th>
                <th className="text-center py-2 text-gray-400 font-medium" colSpan={3}>
                  {myPlayer.username}
                </th>
                {opponentStats && (
                  <th className="text-center py-2 text-gray-400 font-medium" colSpan={3}>
                    {opponent!.username}
                  </th>
                )}
              </tr>
              <tr className="border-b border-gray-600 text-xs text-gray-500">
                <th></th>
                <th className="py-1">Pri</th>
                <th className="py-1">Sec</th>
                <th className="py-1">Gam</th>
                {opponentStats && (
                  <>
                    <th className="py-1">Pri</th>
                    <th className="py-1">Sec</th>
                    <th className="py-1">Gam</th>
                  </>
                )}
              </tr>
            </thead>
            <tbody>
              {rounds.map((r) => {
                const my = myStats.vpByRound[r] ?? { primary: 0, secondary: 0, gambit: 0 };
                const opp = opponentStats?.vpByRound[r] ?? {
                  primary: 0,
                  secondary: 0,
                  gambit: 0,
                };
                return (
                  <tr key={r} className="border-b border-gray-700/50">
                    <td className="py-2 text-gray-400">R{r}</td>
                    <td className="py-2 text-center">{my.primary || "-"}</td>
                    <td className="py-2 text-center">{my.secondary || "-"}</td>
                    <td className="py-2 text-center">{my.gambit || "-"}</td>
                    {opponentStats && (
                      <>
                        <td className="py-2 text-center">{opp.primary || "-"}</td>
                        <td className="py-2 text-center">{opp.secondary || "-"}</td>
                        <td className="py-2 text-center">{opp.gambit || "-"}</td>
                      </>
                    )}
                  </tr>
                );
              })}
              {/* Paint row */}
              <tr className="border-b border-gray-700/50">
                <td className="py-2 text-gray-400">Paint</td>
                <td className="py-2 text-center" colSpan={3}>
                  {myStats.paint}
                </td>
                {opponentStats && (
                  <td className="py-2 text-center" colSpan={3}>
                    {opponentStats.paint}
                  </td>
                )}
              </tr>
              {/* Total row */}
              <tr className="font-bold text-base">
                <td className="py-2 text-gray-300">Total</td>
                <td className="py-2 text-center" colSpan={3}>
                  {myStats.totalVP} VP
                </td>
                {opponentStats && (
                  <td className="py-2 text-center" colSpan={3}>
                    {opponentStats.totalVP} VP
                  </td>
                )}
              </tr>
            </tbody>
          </table>
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
