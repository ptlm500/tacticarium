import { useCallback, useState } from "react";
import { useParams } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useGameStore } from "../stores/gameStore";
import { useGameConnection } from "../hooks/useGameState";
import { getToken } from "../api/client";
import { Stratagem } from "../types/faction";
import { PHASE_LABELS, PHASE_ORDER } from "../types/game";
import { PhaseTracker } from "../components/game/PhaseTracker";
import { RoundIndicator } from "../components/game/RoundIndicator";
import { CPCounter } from "../components/game/CPCounter";
import { VPCounter } from "../components/game/VPCounter";
import { StratagemPanel } from "../components/game/StratagemPanel";
import { SecondaryPanel } from "../components/game/SecondaryPanel";
import { MissionInfo } from "../components/game/MissionInfo";
import { MissionScoring } from "../components/game/MissionScoring";
import { PrimaryScoreHistory } from "../components/game/PrimaryScoreHistory";
import { GameLog } from "../components/game/GameLog";
import { ScoringPrompt, ScoringPromptItem } from "../components/game/ScoringPrompt";
import { PrimaryScoringSlot } from "../types/scoring";
import { ReminderPrompt } from "../components/game/ReminderPrompt";
import { TacticalDrawReminder } from "../components/game/TacticalDrawReminder";
import { ConfirmModal } from "../components/game/ConfirmModal";
import { GameSummary } from "../components/game/GameSummary";
import { useStratagems } from "../hooks/queries/useFactionQueries";
import { useMissions, useMissionRules } from "../hooks/queries/useMissionQueries";

export function GamePage() {
  const { id: gameId } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { gameState, events, error } = useGameStore();

  const token = getToken();

  const { connected, sendAction } = useGameConnection(gameId!, token);

  const [showStratagems, setShowStratagems] = useState(false);
  const [showLog, setShowLog] = useState(false);
  const [showConcedeModal, setShowConcedeModal] = useState(false);
  const [showAbandonModal, setShowAbandonModal] = useState(false);
  const [showRevertModal, setShowRevertModal] = useState(false);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;
  const isMyTurn = myPlayer?.playerNumber === gameState?.activePlayer;

  const [scoringPromptItems, setScoringPromptItems] = useState<ScoringPromptItem[] | null>(null);
  const [showDrawPrompt, setShowDrawPrompt] = useState(false);

  const { data: stratagems = [] } = useStratagems(myPlayer?.factionId);
  const { data: allMissions = [] } = useMissions(gameState?.missionPackId);
  const { data: allRules = [] } = useMissionRules(gameState?.missionPackId);

  const currentMission = allMissions.find((m) => m.id === gameState?.missionId) ?? null;
  const currentTwist = allRules.find((r) => r.id === gameState?.twistId) ?? null;

  // Filter stratagems for current phase
  const availableStratagems = stratagems.filter((s) => {
    if (!gameState) return false;

    const phase = gameState.currentPhase;
    const phaseMatch =
      s.phase === "Any phase" || s.phase.toLowerCase().includes(phase.toLowerCase());

    const turnMatch = isMyTurn
      ? s.turn === "Your turn" || s.turn === "Either player's turn"
      : s.turn === "Opponent's turn" || s.turn === "Either player's turn";

    const detachmentMatch = !s.detachmentId || s.detachmentId === myPlayer?.detachmentId;

    // Challenger stratagems belong to the challenger-card system and are not
    // offered through the general stratagem panel.
    const isChallenger = s.type.startsWith("Challenger \u2013 ");

    return phaseMatch && turnMatch && detachmentMatch && !isChallenger;
  });

  const doAdvancePhase = useCallback(() => {
    setScoringPromptItems(null);
    setShowDrawPrompt(false);
    sendAction("advance_phase");
  }, [sendAction]);

  const handleAdvancePhase = useCallback(() => {
    if (!gameState || !myPlayer) return;

    const phase = gameState.currentPhase;
    const round = gameState.currentRound;
    const isSecondPlayerTurn = gameState.currentTurn === 2;
    const isFightPhase = phase === "fight";
    const isCommandPhase = phase === "command";
    const scoringTiming = currentMission?.scoringTiming || "end_of_command_phase";

    const items: ScoringPromptItem[] = [];

    // Primary scoring prompts
    if (currentMission) {
      if (scoringTiming === "end_of_command_phase") {
        // Prompt when advancing out of Command Phase (BR2+)
        if (isCommandPhase && round >= 2) {
          items.push({
            kind: "primary",
            missionName: currentMission.name,
            scoringRules: (currentMission.scoringRules ?? []).filter(
              (r) => !r.scoringTiming || r.scoringTiming === "end_of_command_phase",
            ),
            currentRound: round,
            scoringSlot: "end_of_command_phase",
          });
        }
        // Also prompt second player at end of turn in BR5
        if (isFightPhase && round === 5 && isSecondPlayerTurn) {
          items.push({
            kind: "primary",
            missionName: currentMission.name,
            scoringRules: (currentMission.scoringRules ?? []).filter(
              (r) => !r.scoringTiming || r.scoringTiming === "end_of_command_phase",
            ),
            currentRound: round,
            scoringSlot: "end_of_command_phase",
          });
        }
      }

      if (scoringTiming === "end_of_battle_round") {
        // Prompt at end of round (second player advancing out of Fight)
        if (isFightPhase && isSecondPlayerTurn) {
          items.push({
            kind: "end_of_round_primary",
            missionName: currentMission.name,
            note: "Both players score at the end of each battle round. Make sure your opponent has scored too.",
          });
        }
      }

      // Per-action end_of_turn scoring (e.g., Terraform bonus)
      if (isFightPhase) {
        const endOfTurnActions = (currentMission.scoringRules ?? []).filter(
          (r) => r.scoringTiming === "end_of_turn",
        );
        if (endOfTurnActions.length > 0) {
          items.push({
            kind: "primary",
            missionName: currentMission.name + " (end of turn)",
            scoringRules: endOfTurnActions,
            currentRound: round,
            scoringSlot: "end_of_turn",
          });
        }
      }
    }

    // Secondary scoring prompt — advancing out of Fight Phase (end of turn)
    if (isFightPhase) {
      if (myPlayer.secondaryMode === "fixed") {
        const fixedSecondaries = (myPlayer.activeSecondaries ?? []).filter((s) => s.isFixed);
        if (fixedSecondaries.length > 0) {
          items.push({ kind: "fixed_secondary", secondaries: fixedSecondaries });
        }
      } else {
        items.push({ kind: "secondary" });
      }
    }

    // Tactical draw prompt — advancing out of Command Phase (separate from scoring)
    let needsDraw = false;
    if (isCommandPhase && myPlayer.secondaryMode === "tactical") {
      const activeCount = myPlayer.activeSecondaries?.length ?? 0;
      const deckSize = myPlayer.tacticalDeck?.length ?? 0;
      if (activeCount < 2 && deckSize > 0) {
        needsDraw = true;
      }
    }

    if (items.length > 0) {
      setScoringPromptItems(items);
    } else if (needsDraw) {
      setShowDrawPrompt(true);
    } else {
      doAdvancePhase();
    }
  }, [gameState, myPlayer, currentMission, doAdvancePhase]);

  const handleAdjustCP = useCallback(
    (delta: number) => {
      sendAction("adjust_cp", { delta });
    },
    [sendAction],
  );

  const handleScoreVP = useCallback(
    (category: string, delta: number, scoringSlot?: PrimaryScoringSlot) => {
      const data: Record<string, unknown> = { category, delta };
      if (scoringSlot) data.scoringSlot = scoringSlot;
      sendAction("score_vp", data);
    },
    [sendAction],
  );

  const handleAdjustVPManual = useCallback(
    (category: string, delta: number) => {
      sendAction("adjust_vp_manual", { category, delta });
    },
    [sendAction],
  );

  const handleUndoPrimaryScore = useCallback(
    (round: number, scoringSlot: PrimaryScoringSlot) => {
      sendAction("undo_primary_score", { round, scoringSlot });
    },
    [sendAction],
  );

  const handleUseStratagem = useCallback(
    (stratagem: Stratagem, cpSpent: number) => {
      sendAction("use_stratagem", {
        stratagemId: stratagem.id,
        cpCost: cpSpent,
      });
    },
    [sendAction],
  );

  const handleConcede = useCallback(() => {
    sendAction("concede");
    setShowConcedeModal(false);
  }, [sendAction]);

  const handleRevertPhase = useCallback(() => {
    sendAction("revert_phase");
    setShowRevertModal(false);
  }, [sendAction]);

  const handleRequestAbandon = useCallback(() => {
    sendAction("request_abandon");
    setShowAbandonModal(false);
  }, [sendAction]);

  const handleRespondAbandon = useCallback(
    (accept: boolean) => {
      sendAction("respond_abandon", { accept });
    },
    [sendAction],
  );

  const handleAchieveSecondary = useCallback(
    (secondaryId: string, vpScored: number) => {
      sendAction("achieve_secondary", { secondaryId, vpScored });
    },
    [sendAction],
  );

  const handleDiscardSecondary = useCallback(
    (secondaryId: string, free: boolean) => {
      sendAction("discard_secondary", { secondaryId, free });
    },
    [sendAction],
  );

  const handleNewOrders = useCallback(
    (discardSecondaryId: string) => {
      sendAction("new_orders", { discardSecondaryId });
    },
    [sendAction],
  );

  const handleReshuffleSecondary = useCallback(
    (secondaryId: string) => {
      sendAction("reshuffle_secondary", { secondaryId });
    },
    [sendAction],
  );

  const handleDrawSecondary = useCallback(() => {
    sendAction("draw_secondary");
  }, [sendAction]);

  const handleDrawChallengerCard = useCallback(() => {
    sendAction("draw_challenger_card", {
      challengerCardId: "challenger-card-generic",
      challengerCardName: "Challenger Mission",
    });
  }, [sendAction]);

  const handleScoreChallenger = useCallback(() => {
    sendAction("score_challenger", {});
  }, [sendAction]);

  if (!gameState || !myPlayer) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>{connected ? "Loading game..." : "Connecting..."}</p>
      </div>
    );
  }

  const totalVP = myPlayer.vpPrimary + myPlayer.vpSecondary + myPlayer.vpGambit + myPlayer.vpPaint;
  const opponentVP = opponent
    ? opponent.vpPrimary + opponent.vpSecondary + opponent.vpGambit + opponent.vpPaint
    : 0;

  if (gameState.status === "completed" || gameState.status === "abandoned") {
    return (
      <GameSummary
        gameState={gameState}
        myPlayer={myPlayer}
        opponent={opponent}
        currentUserId={user?.id ?? ""}
      />
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col">
      {/* Turn Banner */}
      <div
        className={`px-4 py-3 text-center font-semibold ${
          isMyTurn ? "bg-indigo-900" : "bg-gray-800"
        }`}
      >
        Battle Round {gameState.currentRound} — {isMyTurn ? "Your" : `${opponent?.username}'s`} Turn
        — {PHASE_LABELS[gameState.currentPhase]} Phase
      </div>

      {/* Error Banners */}
      {error && <div className="bg-red-900/50 text-red-200 text-center py-2 text-sm">{error}</div>}

      {/* Round & Phase */}
      <div className="px-4 py-3 space-y-2 border-b border-gray-800">
        <RoundIndicator
          currentRound={gameState.currentRound}
          currentTurn={gameState.currentTurn}
          maxRounds={5}
        />
        <PhaseTracker currentPhase={gameState.currentPhase} phases={PHASE_ORDER} />
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Your State */}
        <section className="bg-gray-800 rounded-lg p-4 space-y-3">
          <h2 className="text-sm font-semibold text-gray-400">
            {myPlayer.username} — {myPlayer.factionName}
          </h2>
          <div className="grid grid-cols-2 gap-4">
            <CPCounter
              cp={myPlayer.cp}
              canGainCP={myPlayer.cpGainedThisRound < 1}
              onAdjust={handleAdjustCP}
            />
            <VPCounter
              vpPrimary={myPlayer.vpPrimary}
              vpSecondary={myPlayer.vpSecondary}
              vpGambit={myPlayer.vpGambit}
              vpPaint={myPlayer.vpPaint}
              onAdjust={handleAdjustVPManual}
            />
          </div>
          {/* Mission Quick Scoring */}
          {currentMission &&
            currentMission.scoringRules &&
            currentMission.scoringRules.length > 0 && (
              <MissionScoring
                scoringRules={currentMission.scoringRules ?? []}
                currentRound={gameState.currentRound}
                missionScoringTiming={currentMission.scoringTiming ?? "end_of_command_phase"}
                onScore={(vp, slot) => handleScoreVP("primary", vp, slot)}
              />
            )}
          <PrimaryScoreHistory
            scoredSlots={myPlayer.vpPrimaryScoredSlots ?? {}}
            onUndo={handleUndoPrimaryScore}
          />
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
          currentPhase={gameState.currentPhase}
          isMyTurn={isMyTurn}
          currentCP={myPlayer.cp}
          canGainCP={myPlayer.cpGainedThisRound < 1}
          onAchieve={handleAchieveSecondary}
          onDiscard={handleDiscardSecondary}
          onNewOrders={handleNewOrders}
          onReshuffle={handleReshuffleSecondary}
          onDraw={handleDrawSecondary}
          onScoreFixedVP={(delta) => handleScoreVP("secondary", delta)}
        />

        {/* Challenger Card Banner — only during Command Phase */}
        {opponent &&
          gameState.currentPhase === "command" &&
          totalVP + 6 <= opponentVP &&
          !myPlayer.isChallenger && (
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
            <p className="text-sm text-purple-200 mb-2">Active Challenger Card</p>
            <button
              onClick={handleScoreChallenger}
              className="bg-purple-600 hover:bg-purple-500 text-white font-semibold px-4 py-2 rounded-lg text-sm transition-colors"
            >
              Complete Mission (+3 VP)
            </button>
          </div>
        )}

        {/* Mission Info */}
        <MissionInfo mission={currentMission} twist={currentTwist} />

        {/* Stratagem Panel (collapsible) */}
        <section>
          <button
            onClick={() => setShowStratagems(!showStratagems)}
            className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
          >
            <span className="font-semibold">
              Stratagems ({availableStratagems.length} available)
            </span>
            <span className="text-gray-400">{showStratagems ? "▲" : "▼"}</span>
          </button>
          {showStratagems && (
            <StratagemPanel
              stratagems={availableStratagems}
              currentCP={myPlayer.cp}
              usedThisPhase={myPlayer.stratagemsUsedThisPhase ?? []}
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
            <span className="text-gray-400">{showLog ? "▲" : "▼"}</span>
          </button>
          {showLog && <GameLog events={events} />}
        </section>
      </div>

      {/* Bottom Action Bar */}
      <div className="p-4 border-t border-gray-800 flex gap-3">
        {isMyTurn && (
          <>
            <button
              onClick={() => setShowRevertModal(true)}
              className="bg-gray-700 hover:bg-gray-600 text-gray-200 font-semibold px-4 py-3 rounded-lg transition-colors"
              title="Step back one phase"
            >
              ← Revert
            </button>
            <button
              onClick={handleAdvancePhase}
              className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-3 rounded-lg transition-colors"
            >
              Advance Phase
            </button>
          </>
        )}
        <button
          onClick={() => setShowConcedeModal(true)}
          className="bg-red-900 hover:bg-red-800 text-red-200 font-semibold px-4 py-3 rounded-lg transition-colors"
        >
          Concede
        </button>
        <button
          onClick={() => setShowAbandonModal(true)}
          className="bg-gray-700 hover:bg-gray-600 text-gray-300 font-semibold px-4 py-3 rounded-lg transition-colors"
        >
          Abandon
        </button>
      </div>

      {/* Revert Phase Confirmation Modal */}
      {showRevertModal && (
        <ConfirmModal
          title="Revert Phase"
          message="Step back one phase. If this rolls back into the previous turn, both players lose the 1 CP they gained at the start of this Command phase (clamped at 0). Scored VP, used stratagems, and secondary draws are not reverted."
          confirmLabel="Revert"
          cancelLabel="Cancel"
          variant="default"
          onConfirm={handleRevertPhase}
          onCancel={() => setShowRevertModal(false)}
        />
      )}

      {/* Concede Confirmation Modal */}
      {showConcedeModal && (
        <ConfirmModal
          title="Concede Game"
          message="Are you sure you want to concede? Your opponent will be declared the winner."
          confirmLabel="Concede"
          cancelLabel="Cancel"
          variant="danger"
          onConfirm={handleConcede}
          onCancel={() => setShowConcedeModal(false)}
        />
      )}

      {/* Abandon Request Modal */}
      {showAbandonModal && (
        <ConfirmModal
          title="Abandon Game"
          message="Request to abandon this game with no winner. Your opponent must agree for the game to be abandoned."
          confirmLabel="Request Abandon"
          cancelLabel="Cancel"
          variant="default"
          onConfirm={handleRequestAbandon}
          onCancel={() => setShowAbandonModal(false)}
        />
      )}

      {/* Abandon Request Received Modal */}
      {gameState.abandonRequestedBy != null &&
        gameState.abandonRequestedBy !== myPlayer.playerNumber && (
          <ConfirmModal
            title="Abandon Request"
            message={`${opponent?.username ?? "Your opponent"} wants to abandon this game (no winner). Do you agree?`}
            confirmLabel="Accept"
            cancelLabel="Decline"
            variant="default"
            onConfirm={() => handleRespondAbandon(true)}
            onCancel={() => handleRespondAbandon(false)}
          />
        )}

      {/* Scoring Prompt Modal */}
      {scoringPromptItems && (
        <ScoringPrompt
          items={scoringPromptItems}
          onScore={handleScoreVP}
          activeSecondaries={myPlayer.activeSecondaries ?? []}
          onAchieveSecondary={handleAchieveSecondary}
          onDiscardSecondary={handleDiscardSecondary}
          canGainCP={myPlayer.cpGainedThisRound < 1}
          onScoreFixedVP={(delta) => handleScoreVP("secondary", delta)}
          onConfirm={doAdvancePhase}
          onCancel={() => setScoringPromptItems(null)}
        />
      )}

      {/* Draw Prompt Modal */}
      {showDrawPrompt && (
        <ReminderPrompt
          title="Command Phase Reminder"
          description="Before advancing, check if you need to draw secondaries."
          confirmLabel="Continue"
          cancelLabel="Let me draw first"
          onConfirm={doAdvancePhase}
          onCancel={() => setShowDrawPrompt(false)}
        >
          <TacticalDrawReminder
            deckSize={myPlayer.tacticalDeck?.length ?? 0}
            activeCount={myPlayer.activeSecondaries?.length ?? 0}
            onDraw={handleDrawSecondary}
          />
        </ReminderPrompt>
      )}
    </div>
  );
}
