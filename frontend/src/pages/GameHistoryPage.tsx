import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { gamesApi } from "../api/games";
import { GameSummary } from "../types/game";
import { useAuth } from "../hooks/useAuth";

export function GameHistoryPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [games, setGames] = useState<GameSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    gamesApi
      .getHistory()
      .then(setGames)
      .catch(() => setError("Failed to load game history"))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">Game History</h1>
        <button onClick={() => navigate("/")} className="text-gray-400 hover:text-white">
          Back
        </button>
      </header>

      <main className="max-w-md mx-auto p-6">
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
              const isDraw = !game.winnerId;
              return (
                <div key={game.id} className="bg-gray-800 border border-gray-700 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-medium">{game.missionName || "Unknown Mission"}</span>
                    <span
                      className={`text-xs px-2 py-1 rounded font-semibold ${
                        isDraw
                          ? "bg-gray-700 text-gray-300"
                          : isWinner
                            ? "bg-green-900 text-green-300"
                            : "bg-red-900 text-red-300"
                      }`}
                    >
                      {isDraw ? "Draw" : isWinner ? "Won" : "Lost"}
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
                </div>
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
}
