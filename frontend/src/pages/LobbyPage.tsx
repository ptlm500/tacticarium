import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { gamesApi } from '../api/games';
import { GameSummary } from '../types/game';

export function LobbyPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [games, setGames] = useState<GameSummary[]>([]);
  const [joinCode, setJoinCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    gamesApi.list().then((g) => setGames(g || [])).catch(() => {
      setError('Failed to load your games');
    });
  }, []);

  const handleCreate = async () => {
    setLoading(true);
    try {
      const { id } = await gamesApi.create();
      navigate(`/game/${id}/setup`);
    } catch {
      setError('Failed to create game');
    } finally {
      setLoading(false);
    }
  };

  const handleJoin = async () => {
    if (!joinCode.trim()) return;
    setLoading(true);
    try {
      const { id } = await gamesApi.join(joinCode.trim().toUpperCase());
      navigate(`/game/${id}/setup`);
    } catch {
      setError('Invalid invite code');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">Tacticarium</h1>
        <div className="flex items-center gap-4">
          <span className="text-gray-400">{user?.username}</span>
          <button
            onClick={() => navigate('/history')}
            className="text-gray-400 hover:text-white text-sm"
          >
            History
          </button>
          <button
            onClick={logout}
            className="text-gray-400 hover:text-white text-sm"
          >
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
                <button
                  key={game.id}
                  onClick={() =>
                    navigate(
                      game.status === 'setup'
                        ? `/game/${game.id}/setup`
                        : `/game/${game.id}`
                    )
                  }
                  className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg p-4 text-left transition-colors"
                >
                  <div className="flex justify-between items-center">
                    <span className="font-medium">
                      {game.missionName || 'No mission selected'}
                    </span>
                    <span
                      className={`text-xs px-2 py-1 rounded ${
                        game.status === 'active'
                          ? 'bg-green-900 text-green-300'
                          : game.status === 'setup'
                          ? 'bg-yellow-900 text-yellow-300'
                          : 'bg-gray-700 text-gray-400'
                      }`}
                    >
                      {game.status}
                    </span>
                  </div>
                  <div className="text-sm text-gray-400 mt-1">
                    {game.players.map((p) => p.username).join(' vs ')}
                  </div>
                </button>
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
