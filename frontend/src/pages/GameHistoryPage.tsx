import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { gamesApi } from "../api/games";
import { factionsApi } from "../api/factions";
import { GameSummary, UserStats } from "../types/game";
import { Faction } from "../types/faction";
import { useAuth } from "../hooks/useAuth";
import { ConfirmModal } from "../components/game/ConfirmModal";

export function GameHistoryPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [games, setGames] = useState<GameSummary[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [factions, setFactions] = useState<Faction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [myFaction, setMyFaction] = useState("");
  const [opponentFaction, setOpponentFaction] = useState("");
  const [gameToRemove, setGameToRemove] = useState<GameSummary | null>(null);

  // Load stats and factions once
  useEffect(() => {
    gamesApi
      .getStats()
      .then(setStats)
      .catch(() => {});
    factionsApi
      .list()
      .then(setFactions)
      .catch(() => {});
  }, []);

  // Load history (re-fetch when filters change)
  useEffect(() => {
    setLoading(true);
    const filters: { myFaction?: string; opponentFaction?: string } = {};
    if (myFaction) filters.myFaction = myFaction;
    if (opponentFaction) filters.opponentFaction = opponentFaction;

    gamesApi
      .getHistory(Object.keys(filters).length > 0 ? filters : undefined)
      .then(setGames)
      .catch(() => setError("Failed to load game history"))
      .finally(() => setLoading(false));
  }, [myFaction, opponentFaction]);

  const handleRemoveGame = async () => {
    if (!gameToRemove) return;
    try {
      await gamesApi.hide(gameToRemove.id);
      setGames((prev) => prev.filter((g) => g.id !== gameToRemove.id));
    } catch {
      setError("Failed to remove game");
    } finally {
      setGameToRemove(null);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">Game History</h1>
        <button onClick={() => navigate("/")} className="text-gray-400 hover:text-white">
          Back
        </button>
      </header>

      <main className="max-w-md mx-auto p-6 space-y-6">
        {/* Stats Summary */}
        {stats && (
          <section className="bg-gray-800 border border-gray-700 rounded-lg p-4 space-y-3">
            <div className="flex flex-wrap gap-3 justify-center">
              <span className="text-xs px-2 py-1 rounded font-semibold bg-green-900 text-green-300">
                {stats.wins}W
              </span>
              <span className="text-xs px-2 py-1 rounded font-semibold bg-red-900 text-red-300">
                {stats.losses}L
              </span>
              <span className="text-xs px-2 py-1 rounded font-semibold bg-gray-700 text-gray-300">
                {stats.draws}D
              </span>
              {stats.abandoned > 0 && (
                <span className="text-xs px-2 py-1 rounded font-semibold bg-yellow-900 text-yellow-300">
                  {stats.abandoned} Abandoned
                </span>
              )}
            </div>
            <div className="text-center text-sm text-gray-400">
              Avg VP: <span className="text-white">{stats.averageVp.toFixed(1)}</span>
            </div>
            {(stats.factionStats ?? []).length > 0 && (
              <div className="text-xs text-gray-400 space-y-1">
                <p className="font-medium text-gray-500">Most played:</p>
                {(stats.factionStats ?? []).slice(0, 3).map((fs) => (
                  <div key={fs.factionName} className="flex justify-between">
                    <span>{fs.factionName}</span>
                    <span>
                      {fs.gamesPlayed} game{fs.gamesPlayed !== 1 ? "s" : ""}, {fs.wins} win
                      {fs.wins !== 1 ? "s" : ""}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </section>
        )}

        {/* Filters */}
        <div className="flex gap-3">
          <select
            value={myFaction}
            onChange={(e) => setMyFaction(e.target.value)}
            className="flex-1 bg-gray-800 border border-gray-700 text-gray-300 text-sm px-3 py-2 rounded-lg"
          >
            <option value="">My Faction (all)</option>
            {factions.map((f) => (
              <option key={f.id} value={f.name}>
                {f.name}
              </option>
            ))}
          </select>
          <select
            value={opponentFaction}
            onChange={(e) => setOpponentFaction(e.target.value)}
            className="flex-1 bg-gray-800 border border-gray-700 text-gray-300 text-sm px-3 py-2 rounded-lg"
          >
            <option value="">Opponent (all)</option>
            {factions.map((f) => (
              <option key={f.id} value={f.name}>
                {f.name}
              </option>
            ))}
          </select>
        </div>

        {/* Game List */}
        {error ? (
          <div className="bg-red-900/50 border border-red-700 text-red-200 px-4 py-2 rounded text-center">
            {error}
          </div>
        ) : loading ? (
          <p className="text-gray-500 text-center">Loading...</p>
        ) : games.length === 0 ? (
          <p className="text-gray-500 text-center">No completed games yet.</p>
        ) : (
          <div className="space-y-3">
            {games.map((game) => {
              const isWinner = game.winnerId === user?.id;
              const isDraw = !game.winnerId && game.status === "completed";
              const isAbandoned = game.status === "abandoned";
              return (
                <div key={game.id} className="flex gap-2">
                  <button
                    onClick={() => navigate(`/history/${game.id}`)}
                    className="flex-1 text-left bg-gray-800 border border-gray-700 rounded-lg p-4 hover:border-gray-600 transition-colors"
                  >
                    <div className="flex justify-between items-center mb-2">
                      <span className="font-medium">{game.missionName || "Unknown Mission"}</span>
                      <span
                        className={`text-xs px-2 py-1 rounded font-semibold ${
                          isAbandoned
                            ? "bg-yellow-900 text-yellow-300"
                            : isDraw
                              ? "bg-gray-700 text-gray-300"
                              : isWinner
                                ? "bg-green-900 text-green-300"
                                : "bg-red-900 text-red-300"
                        }`}
                      >
                        {isAbandoned ? "Abandoned" : isDraw ? "Draw" : isWinner ? "Won" : "Lost"}
                      </span>
                    </div>
                    <div className="space-y-1 text-sm">
                      {(game.players ?? []).map((p) => (
                        <div
                          key={p.userId}
                          className={`flex justify-between ${
                            p.userId === user?.id ? "text-white" : "text-gray-400"
                          }`}
                        >
                          <span>
                            {p.username}
                            {p.factionName && ` (${p.factionName})`}
                          </span>
                          <span>{p.totalVp} VP</span>
                        </div>
                      ))}
                    </div>
                    {game.completedAt && (
                      <p className="text-xs text-gray-500 mt-2">
                        {new Date(game.completedAt).toLocaleDateString()}
                      </p>
                    )}
                  </button>
                  <button
                    onClick={() => setGameToRemove(game)}
                    className="self-center px-3 py-2 text-gray-500 hover:text-red-400 transition-colors"
                    aria-label="Remove game"
                    title="Remove game"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-5 w-5"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </main>

      {gameToRemove && (
        <ConfirmModal
          title="Remove Game"
          message="Are you sure you want to remove this game from your history? This cannot be undone."
          confirmLabel="Remove"
          cancelLabel="Cancel"
          variant="danger"
          onConfirm={handleRemoveGame}
          onCancel={() => setGameToRemove(null)}
        />
      )}
    </div>
  );
}
