import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useGameStore } from '../stores/gameStore';
import { useGameConnection } from '../hooks/useGameState';
import { factionsApi } from '../api/factions';
import { Faction, Detachment } from '../types/faction';
import { FactionPicker } from '../components/setup/FactionPicker';
import { DetachmentPicker } from '../components/setup/DetachmentPicker';

export function GameSetupPage() {
  const { id: gameId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { gameState } = useGameStore();

  // Get token from cookie for WS connection
  const token = document.cookie
    .split('; ')
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] || '';

  const { connected, sendAction } = useGameConnection(gameId!, token);

  const [factions, setFactions] = useState<Faction[]>([]);
  const [detachments, setDetachments] = useState<Detachment[]>([]);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    factionsApi.list().then(setFactions);
  }, []);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;

  // Load detachments when faction changes
  useEffect(() => {
    if (myPlayer?.factionId) {
      factionsApi.getDetachments(myPlayer.factionId).then(setDetachments);
    } else {
      setDetachments([]);
    }
  }, [myPlayer?.factionId]);

  // Navigate to game when it starts
  useEffect(() => {
    if (gameState?.status === 'active') {
      navigate(`/game/${gameId}`);
    }
  }, [gameState?.status, gameId, navigate]);

  const handleSelectFaction = useCallback(
    (faction: Faction) => {
      sendAction('select_faction', {
        factionId: faction.id,
        factionName: faction.name,
      });
    },
    [sendAction]
  );

  const handleSelectDetachment = useCallback(
    (detachment: Detachment) => {
      sendAction('select_detachment', {
        detachmentId: detachment.id,
        detachmentName: detachment.name,
      });
    },
    [sendAction]
  );

  const handleReady = useCallback(() => {
    sendAction('set_ready', { ready: !myPlayer?.ready });
  }, [sendAction, myPlayer?.ready]);

  const copyInviteCode = () => {
    if (gameState?.inviteCode) {
      navigator.clipboard.writeText(gameState.inviteCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!gameState) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>{connected ? 'Loading game...' : 'Connecting...'}</p>
      </div>
    );
  }

  const canReady = myPlayer?.factionId && myPlayer?.detachmentId;

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="p-4 border-b border-gray-800">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">Game Setup</h1>
          <button
            onClick={copyInviteCode}
            className="bg-gray-800 hover:bg-gray-700 border border-gray-600 px-4 py-2 rounded-lg text-sm transition-colors"
          >
            {copied ? 'Copied!' : `Invite: ${gameState.inviteCode}`}
          </button>
        </div>
        {!opponent && (
          <p className="text-yellow-400 text-sm mt-2">
            Waiting for opponent to join...
          </p>
        )}
      </header>

      <main className="max-w-md mx-auto p-6 space-y-6">
        {/* Faction Selection */}
        <section>
          <h2 className="text-lg font-semibold mb-3">Your Faction</h2>
          <FactionPicker
            factions={factions}
            selectedId={myPlayer?.factionId || ''}
            onSelect={handleSelectFaction}
          />
        </section>

        {/* Detachment Selection */}
        {myPlayer?.factionId && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Detachment</h2>
            <DetachmentPicker
              detachments={detachments}
              selectedId={myPlayer?.detachmentId || ''}
              onSelect={handleSelectDetachment}
            />
          </section>
        )}

        {/* Opponent Status */}
        {opponent && (
          <section className="bg-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-semibold text-gray-400 mb-2">
              Opponent: {opponent.username}
            </h2>
            <p className="text-sm">
              {opponent.factionName || 'Selecting faction...'}
              {opponent.detachmentName && ` - ${opponent.detachmentName}`}
            </p>
            <p className="text-sm mt-1">
              {opponent.ready ? (
                <span className="text-green-400">Ready</span>
              ) : (
                <span className="text-yellow-400">Not ready</span>
              )}
            </p>
          </section>
        )}

        {/* Ready Button */}
        <button
          onClick={handleReady}
          disabled={!canReady}
          className={`w-full font-semibold py-3 rounded-lg transition-colors ${
            myPlayer?.ready
              ? 'bg-green-700 hover:bg-green-800 text-white'
              : 'bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white'
          }`}
        >
          {myPlayer?.ready ? 'Ready! (click to unready)' : 'Ready Up'}
        </button>
      </main>
    </div>
  );
}
