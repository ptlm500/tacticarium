import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { GameSummary } from "../types/game";
import { ConfirmModal } from "../components/game/ConfirmModal";
import { useGameList } from "../hooks/queries/useGamesQueries";
import { useCreateGame, useJoinGame, useHideGame } from "../hooks/queries/useGameMutations";

export function LobbyPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [joinCode, setJoinCode] = useState("");
  const [gameToRemove, setGameToRemove] = useState<GameSummary | null>(null);

  const { data: games = [] } = useGameList();
  const createGame = useCreateGame();
  const joinGame = useJoinGame();
  const hideGame = useHideGame();

  const loading = createGame.isPending || joinGame.isPending;
  const error = createGame.error
    ? "Failed to create game"
    : joinGame.error
      ? "Invalid invite code"
      : hideGame.error
        ? "Failed to remove game"
        : "";

  const handleCreate = () => {
    createGame.mutate(undefined, {
      onSuccess: ({ id }) => navigate(`/game/${id}/setup`),
    });
  };

  const handleJoin = () => {
    if (!joinCode.trim()) return;
    joinGame.mutate(joinCode.trim().toUpperCase(), {
      onSuccess: ({ id }) => navigate(`/game/${id}/setup`),
    });
  };

  const handleRemoveGame = () => {
    if (!gameToRemove) return;
    hideGame.mutate(gameToRemove.id, {
      onSettled: () => setGameToRemove(null),
    });
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">Tacticarium</h1>
        <div className="flex items-center gap-4">
          <span className="text-gray-400">{user?.username}</span>
          <button
            onClick={() => navigate("/history")}
            className="text-gray-400 hover:text-white text-sm"
          >
            History
          </button>
          <button onClick={logout} className="text-gray-400 hover:text-white text-sm">
            Logout
          </button>
        </div>
      </header>

      <main className="max-w-md mx-auto p-6 space-y-8">
        {error && (
          <div className="bg-red-900/50 border border-red-700 text-red-200 px-4 py-2 rounded">
            {error}
          </div>
        )}

        <section className="space-y-4">
          <h2 className="text-lg font-semibold">New Game</h2>
          <button
            onClick={handleCreate}
            disabled={loading}
            className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
          >
            Create Game
          </button>
        </section>

        <section className="space-y-4">
          <h2 className="text-lg font-semibold">Join Game</h2>
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="Enter invite code"
              value={joinCode}
              onChange={(e) => setJoinCode(e.target.value)}
              className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
              maxLength={6}
            />
            <button
              onClick={handleJoin}
              disabled={loading || !joinCode.trim()}
              className="bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-semibold px-6 py-2 rounded-lg transition-colors"
            >
              Join
            </button>
          </div>
        </section>

        {games.length > 0 && (
          <section className="space-y-4">
            <h2 className="text-lg font-semibold">Your Games</h2>
            <div className="space-y-2">
              {games.map((game) => (
                <div key={game.id} className="flex gap-2">
                  <button
                    onClick={() =>
                      navigate(
                        game.status === "setup" ? `/game/${game.id}/setup` : `/game/${game.id}`,
                      )
                    }
                    className="flex-1 bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg p-4 text-left transition-colors"
                  >
                    <div className="flex justify-between items-center">
                      <span className="font-medium">
                        {game.missionName || "No mission selected"}
                      </span>
                      <span
                        className={`text-xs px-2 py-1 rounded ${
                          game.status === "active"
                            ? "bg-green-900 text-green-300"
                            : game.status === "setup"
                              ? "bg-yellow-900 text-yellow-300"
                              : "bg-gray-700 text-gray-400"
                        }`}
                      >
                        {game.status}
                      </span>
                    </div>
                    <div className="text-sm text-gray-400 mt-1">
                      {(game.players ?? []).map((p) => p.username).join(" vs ")}
                    </div>
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
              ))}
            </div>
          </section>
        )}
      </main>

      {gameToRemove && (
        <ConfirmModal
          title="Remove Game"
          message="Are you sure you want to remove this game? It will no longer appear in your game list. This cannot be undone."
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
