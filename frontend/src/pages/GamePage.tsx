import { useEffect, useCallback, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useGameStore } from '../stores/gameStore';
import { useGameConnection } from '../hooks/useGameState';
import { getToken } from '../api/client';
import { factionsApi } from '../api/factions';
import { missionsApi } from '../api/missions';
import { Stratagem } from '../types/faction';
import { Mission } from '../types/mission';
import { PHASE_LABELS, PHASE_ORDER } from '../types/game';
import { PhaseTracker } from '../components/game/PhaseTracker';
import { RoundIndicator } from '../components/game/RoundIndicator';
import { CPCounter } from '../components/game/CPCounter';
import { VPCounter } from '../components/game/VPCounter';
import { StratagemPanel } from '../components/game/StratagemPanel';
import { SecondaryPanel } from '../components/game/SecondaryPanel';
import { MissionInfo } from '../components/game/MissionInfo';
import { MissionScoring } from '../components/game/MissionScoring';
import { GameLog } from '../components/game/GameLog';
import { ScoringPrompt, ScoringPromptItem } from '../components/game/ScoringPrompt';

export function GamePage() {
  const { id: gameId } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { gameState, events, error } = useGameStore();

  const token = getToken();

  const { connected, sendAction } = useGameConnection(gameId!, token);

  const [stratagems, setStratagems] = useState<Stratagem[]>([]);
  const [currentMission, setCurrentMission] = useState<Mission | null>(null);
  const [showStratagems, setShowStratagems] = useState(false);
  const [showLog, setShowLog] = useState(false);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;
  const isMyTurn = myPlayer?.playerNumber === gameState?.activePlayer;

  const [loadError, setLoadError] = useState('');
  const [scoringPromptItems, setScoringPromptItems] = useState<ScoringPromptItem[] | null>(null);

  // Load stratagems for player's faction
  useEffect(() => {
    if (myPlayer?.factionId) {
      factionsApi.getStratagems(myPlayer.factionId).then(setStratagems)
        .catch(() => setLoadError('Failed to load stratagems'));
    }
  }, [myPlayer?.factionId]);

  // Load mission scoring rules
  useEffect(() => {
    if (gameState?.missionPackId && gameState?.missionId) {
      missionsApi.listMissions(gameState.missionPackId).then((missions) => {
        const m = missions.find((m) => m.id === gameState.missionId);
        setCurrentMission(m ?? null);
      }).catch(() => setLoadError('Failed to load mission data'));
    }
  }, [gameState?.missionPackId, gameState?.missionId]);

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

  const doAdvancePhase = useCallback(() => {
    setScoringPromptItems(null);
    sendAction('advance_phase');
  }, [sendAction]);

  const handleAdvancePhase = useCallback(() => {
    if (!gameState || !myPlayer) return;

    const phase = gameState.currentPhase;
    const round = gameState.currentRound;
    const isSecondPlayerTurn = gameState.currentTurn === 2;
    const isFightPhase = phase === 'fight';
    const isCommandPhase = phase === 'command';
    const scoringTiming = currentMission?.scoringTiming || 'end_of_command_phase';

    const items: ScoringPromptItem[] = [];

    // Primary scoring prompts
    if (currentMission) {
      if (scoringTiming === 'end_of_command_phase') {
        // Prompt when advancing out of Command Phase (BR2+)
        if (isCommandPhase && round >= 2) {
          items.push({
            kind: 'primary',
            missionName: currentMission.name,
            scoringRules: currentMission.scoringRules.filter(
              (r) => !r.scoringTiming || r.scoringTiming === 'end_of_command_phase'
            ),
            currentRound: round,
          });
        }
        // Also prompt second player at end of turn in BR5
        if (isFightPhase && round === 5 && isSecondPlayerTurn) {
          items.push({
            kind: 'primary',
            missionName: currentMission.name,
            scoringRules: currentMission.scoringRules.filter(
              (r) => !r.scoringTiming || r.scoringTiming === 'end_of_command_phase'
            ),
            currentRound: round,
          });
        }
      }

      if (scoringTiming === 'end_of_battle_round') {
        // Prompt at end of round (second player advancing out of Fight)
        if (isFightPhase && isSecondPlayerTurn) {
          items.push({
            kind: 'end_of_round_primary',
            missionName: currentMission.name,
            note: 'Both players score at the end of each battle round. Make sure your opponent has scored too.',
          });
        }
      }

      // Per-action end_of_turn scoring (e.g., Terraform bonus)
      if (isFightPhase) {
        const endOfTurnActions = currentMission.scoringRules.filter(
          (r) => r.scoringTiming === 'end_of_turn'
        );
        if (endOfTurnActions.length > 0) {
          items.push({
            kind: 'primary',
            missionName: currentMission.name + ' (end of turn)',
            scoringRules: endOfTurnActions,
            currentRound: round,
          });
        }
      }
    }

    // Tactical draw prompt — advancing out of Command Phase
    if (isCommandPhase && myPlayer.secondaryMode === 'tactical') {
      const activeCount = myPlayer.activeSecondaries?.length ?? 0;
      const deckSize = myPlayer.tacticalDeck?.length ?? 0;
      if (activeCount < 2 && deckSize > 0) {
        items.push({ kind: 'tactical_draw' });
      }
    }

    // Secondary scoring prompt — advancing out of Fight Phase (end of turn)
    if (isFightPhase) {
      if (myPlayer.secondaryMode === 'fixed') {
        const fixedSecondaries = (myPlayer.activeSecondaries ?? []).filter((s) => s.isFixed);
        if (fixedSecondaries.length > 0) {
          items.push({ kind: 'fixed_secondary', secondaries: fixedSecondaries });
        }
      } else {
        items.push({ kind: 'secondary' });
      }
    }

    if (items.length > 0) {
      setScoringPromptItems(items);
    } else {
      doAdvancePhase();
    }
  }, [gameState, myPlayer, currentMission, doAdvancePhase]);

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

  const handleAchieveSecondary = useCallback(
    (secondaryId: string, vpScored: number) => {
      sendAction('achieve_secondary', { secondaryId, vpScored });
    },
    [sendAction]
  );

  const handleDiscardSecondary = useCallback(
    (secondaryId: string, free: boolean) => {
      sendAction('discard_secondary', { secondaryId, free });
    },
    [sendAction]
  );

  const handleNewOrders = useCallback(
    (discardSecondaryId: string) => {
      sendAction('new_orders', { discardSecondaryId });
    },
    [sendAction]
  );

  const handleDrawSecondary = useCallback(() => {
    sendAction('draw_secondary');
  }, [sendAction]);

  const handleDrawChallengerCard = useCallback(() => {
    // For now, pick a placeholder challenger card — the UI could be enhanced
    // with a card picker in the future
    sendAction('draw_challenger_card', {
      challengerCardId: 'challenger-card-generic',
      challengerCardName: 'Challenger Mission',
    });
  }, [sendAction]);

  const handleScoreChallenger = useCallback(() => {
    sendAction('score_challenger', {});
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
        Battle Round {gameState.currentRound} —{' '}
        {isMyTurn ? 'Your' : `${opponent?.username}'s`} Turn —{' '}
        {PHASE_LABELS[gameState.currentPhase]} Phase
      </div>

      {/* Error Banners */}
      {error && (
        <div className="bg-red-900/50 text-red-200 text-center py-2 text-sm">
          {error}
        </div>
      )}
      {loadError && (
        <div className="bg-red-900/50 border border-red-700 text-red-200 text-center py-2 text-sm">
          {loadError}
        </div>
      )}

      {/* Round & Phase */}
      <div className="px-4 py-3 space-y-2 border-b border-gray-800">
        <RoundIndicator currentRound={gameState.currentRound} currentTurn={gameState.currentTurn} maxRounds={5} />
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
            <CPCounter cp={myPlayer.cp} canGainCP={myPlayer.cpGainedThisRound < 1} onAdjust={handleAdjustCP} />
            <VPCounter
              vpPrimary={myPlayer.vpPrimary}
              vpSecondary={myPlayer.vpSecondary}
              vpGambit={myPlayer.vpGambit}
              vpPaint={myPlayer.vpPaint}
              onScore={handleScoreVP}
            />
          </div>
          {/* Mission Quick Scoring */}
          {currentMission && currentMission.scoringRules.length > 0 && (
            <MissionScoring
              scoringRules={currentMission.scoringRules}
              currentRound={gameState.currentRound}
              onScore={(vp) => handleScoreVP('primary', vp)}
            />
          )}
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

        {/* Secondary Missions */}
        <SecondaryPanel
          mode={myPlayer.secondaryMode}
          activeSecondaries={myPlayer.activeSecondaries ?? []}
          achievedSecondaries={myPlayer.achievedSecondaries ?? []}
          discardedSecondaries={myPlayer.discardedSecondaries ?? []}
          deckSize={myPlayer.tacticalDeck?.length ?? 0}
          currentRound={gameState.currentRound}
          currentCP={myPlayer.cp}
          canGainCP={myPlayer.cpGainedThisRound < 1}
          onAchieve={handleAchieveSecondary}
          onDiscard={handleDiscardSecondary}
          onNewOrders={handleNewOrders}
          onDraw={handleDrawSecondary}
          onScoreFixedVP={(delta) => handleScoreVP('secondary', delta)}
        />

        {/* Challenger Card Banner — only during Command Phase */}
        {opponent && gameState.currentPhase === 'command' && totalVP + 6 <= opponentVP && !myPlayer.isChallenger && (
          <div className="bg-amber-900/50 border border-amber-700 rounded-lg p-4 text-center">
            <p className="text-sm text-amber-200 mb-2">
              You are trailing by {opponentVP - totalVP} VP — eligible for a Challenger Card!
            </p>
            <button
              onClick={handleDrawChallengerCard}
              className="bg-amber-600 hover:bg-amber-500 text-white font-semibold px-4 py-2 rounded-lg text-sm transition-colors"
            >
              Draw Challenger Card
            </button>
          </div>
        )}

        {/* Active Challenger Card */}
        {myPlayer.isChallenger && myPlayer.challengerCardId && (
          <div className="bg-purple-900/50 border border-purple-700 rounded-lg p-4">
            <p className="text-sm text-purple-200 mb-2">
              Active Challenger Card
            </p>
            <button
              onClick={handleScoreChallenger}
              className="bg-purple-600 hover:bg-purple-500 text-white font-semibold px-4 py-2 rounded-lg text-sm transition-colors"
            >
              Complete Mission (+3 VP)
            </button>
          </div>
        )}

        {/* Mission Info */}
        <MissionInfo
          missionName={gameState.missionName || ''}
          twistName={gameState.twistName || ''}
        />

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

      {/* Scoring Prompt Modal */}
      {scoringPromptItems && (
        <ScoringPrompt
          items={scoringPromptItems}
          onScore={handleScoreVP}
          activeSecondaries={myPlayer.activeSecondaries ?? []}
          onAchieveSecondary={handleAchieveSecondary}
          onDiscardSecondary={handleDiscardSecondary}
          onDrawSecondary={handleDrawSecondary}
          canGainCP={myPlayer.cpGainedThisRound < 1}
          deckSize={myPlayer.tacticalDeck?.length ?? 0}
          activeSecondaryCount={myPlayer.activeSecondaries?.length ?? 0}
          onScoreFixedVP={(delta) => handleScoreVP('secondary', delta)}
          onConfirm={doAdvancePhase}
          onCancel={() => setScoringPromptItems(null)}
        />
      )}
    </div>
  );
}
