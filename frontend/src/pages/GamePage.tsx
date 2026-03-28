import { useEffect, useCallback, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useGameStore } from '../stores/gameStore';
import { useGameConnection } from '../hooks/useGameState';
import { getToken } from '../api/client';
import { factionsApi } from '../api/factions';
import { Stratagem } from '../types/faction';
import { PHASE_LABELS, PHASE_ORDER } from '../types/game';
import { PhaseTracker } from '../components/game/PhaseTracker';
import { RoundIndicator } from '../components/game/RoundIndicator';
import { CPCounter } from '../components/game/CPCounter';
import { VPCounter } from '../components/game/VPCounter';
import { StratagemPanel } from '../components/game/StratagemPanel';
import { GameLog } from '../components/game/GameLog';

export function GamePage() {
  const { id: gameId } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { gameState, events, error } = useGameStore();

  const token = getToken();

  const { connected, sendAction } = useGameConnection(gameId!, token);

  const [stratagems, setStratagems] = useState<Stratagem[]>([]);
  const [showStratagems, setShowStratagems] = useState(false);
  const [showLog, setShowLog] = useState(false);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;
  const isMyTurn = myPlayer?.playerNumber === gameState?.activePlayer;

  // Load stratagems for player's faction
  useEffect(() => {
    if (myPlayer?.factionId) {
      factionsApi.getStratagems(myPlayer.factionId).then(setStratagems);
    }
  }, [myPlayer?.factionId]);

  // Filter stratagems for current phase
  const availableStratagems = stratagems.filter((s) => {
    if (!gameState) return false;

    const phase = gameState.currentPhase;
    const phaseMatch =
      s.phase === 'Any phase' ||
      s.phase.toLowerCase().includes(phase.toLowerCase());

    const turnMatch = isMyTurn
      ? s.turn === 'Your turn' || s.turn === "Either player's turn"
      : s.turn === "Opponent's turn" || s.turn === "Either player's turn";

    const detachmentMatch =
      !s.detachmentId || s.detachmentId === myPlayer?.detachmentId;

    return phaseMatch && turnMatch && detachmentMatch;
  });

  const handleAdvancePhase = useCallback(() => {
    sendAction('advance_phase');
  }, [sendAction]);

  const handleAdjustCP = useCallback(
    (delta: number) => {
      sendAction('adjust_cp', { delta });
    },
    [sendAction]
  );

  const handleScoreVP = useCallback(
    (category: string, delta: number) => {
      sendAction('score_vp', { category, delta });
    },
    [sendAction]
  );

  const handleUseStratagem = useCallback(
    (stratagem: Stratagem) => {
      sendAction('use_stratagem', {
        stratagemId: stratagem.id,
        stratagemName: stratagem.name,
        cpCost: stratagem.cpCost,
      });
    },
    [sendAction]
  );

  const handleConcede = useCallback(() => {
    if (window.confirm('Are you sure you want to concede?')) {
      sendAction('concede');
    }
  }, [sendAction]);

  if (!gameState || !myPlayer) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>{connected ? 'Loading game...' : 'Connecting...'}</p>
      </div>
    );
  }

  const totalVP = myPlayer.vpPrimary + myPlayer.vpSecondary + myPlayer.vpGambit + myPlayer.vpPaint;
  const opponentVP = opponent
    ? opponent.vpPrimary + opponent.vpSecondary + opponent.vpGambit + opponent.vpPaint
    : 0;

  if (gameState.status === 'completed') {
    const isWinner = gameState.winnerId === user?.id;
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <div className="text-center p-8 space-y-4">
          <h1 className="text-3xl font-bold">
            {gameState.winnerId
              ? isWinner
                ? 'Victory!'
                : 'Defeat'
              : 'Draw'}
          </h1>
          <div className="space-y-2">
            <p className="text-xl">
              {myPlayer.username}: {totalVP} VP
            </p>
            {opponent && (
              <p className="text-xl text-gray-400">
                {opponent.username}: {opponentVP} VP
              </p>
            )}
          </div>
          <a
            href="/"
            className="inline-block bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-6 rounded-lg"
          >
            Back to Lobby
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col">
      {/* Turn Banner */}
      <div
        className={`px-4 py-3 text-center font-semibold ${
          isMyTurn ? 'bg-indigo-900' : 'bg-gray-800'
        }`}
      >
        Round {gameState.currentRound} —{' '}
        {isMyTurn ? 'Your' : `${opponent?.username}'s`}{' '}
        {PHASE_LABELS[gameState.currentPhase]} Phase
      </div>

      {/* Error Banner */}
      {error && (
        <div className="bg-red-900/50 text-red-200 text-center py-2 text-sm">
          {error}
        </div>
      )}

      {/* Round & Phase */}
      <div className="px-4 py-3 space-y-2 border-b border-gray-800">
        <RoundIndicator currentRound={gameState.currentRound} maxRounds={5} />
        <PhaseTracker
          currentPhase={gameState.currentPhase}
          phases={PHASE_ORDER}
        />
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Your State */}
        <section className="bg-gray-800 rounded-lg p-4 space-y-3">
          <h2 className="text-sm font-semibold text-gray-400">
            {myPlayer.username} — {myPlayer.factionName}
          </h2>
          <div className="grid grid-cols-2 gap-4">
            <CPCounter cp={myPlayer.cp} onAdjust={handleAdjustCP} />
            <VPCounter
              vpPrimary={myPlayer.vpPrimary}
              vpSecondary={myPlayer.vpSecondary}
              vpGambit={myPlayer.vpGambit}
              vpPaint={myPlayer.vpPaint}
              onScore={handleScoreVP}
            />
          </div>
        </section>

        {/* Opponent State */}
        {opponent && (
          <section className="bg-gray-800/50 rounded-lg p-4">
            <h2 className="text-sm font-semibold text-gray-400">
              {opponent.username} — {opponent.factionName}
            </h2>
            <div className="flex gap-6 mt-2">
              <span>CP: {opponent.cp}</span>
              <span>VP: {opponentVP}</span>
            </div>
          </section>
        )}

        {/* Stratagem Panel (collapsible) */}
        <section>
          <button
            onClick={() => setShowStratagems(!showStratagems)}
            className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
          >
            <span className="font-semibold">
              Stratagems ({availableStratagems.length} available)
            </span>
            <span className="text-gray-400">{showStratagems ? '▲' : '▼'}</span>
          </button>
          {showStratagems && (
            <StratagemPanel
              stratagems={availableStratagems}
              currentCP={myPlayer.cp}
              onUse={handleUseStratagem}
            />
          )}
        </section>

        {/* Game Log (collapsible) */}
        <section>
          <button
            onClick={() => setShowLog(!showLog)}
            className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
          >
            <span className="font-semibold">Game Log</span>
            <span className="text-gray-400">{showLog ? '▲' : '▼'}</span>
          </button>
          {showLog && <GameLog events={events} />}
        </section>
      </div>

      {/* Bottom Action Bar */}
      <div className="p-4 border-t border-gray-800 flex gap-3">
        {isMyTurn && (
          <button
            onClick={handleAdvancePhase}
            className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-3 rounded-lg transition-colors"
          >
            Advance Phase
          </button>
        )}
        <button
          onClick={handleConcede}
          className="bg-red-900 hover:bg-red-800 text-red-200 font-semibold px-4 py-3 rounded-lg transition-colors"
        >
          Concede
        </button>
      </div>
    </div>
  );
}
