import { useCallback, useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ChevronDown, ChevronUp, Flag, Forward, Handshake, ScrollText, Zap } from "lucide-react";
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
import { BoardView } from "../components/game/BoardView";
import { ScorePromptsPanel } from "../components/game/ScorePromptsPanel";
import { MissionInfo } from "../components/game/MissionInfo";
import { GameLog } from "../components/game/GameLog";
import { PlayerAvatar } from "../components/game/PlayerAvatar";
import { ConfirmModal } from "../components/game/ConfirmModal";
import { GameSummary } from "../components/game/GameSummary";
import { SecondaryDetailsModal } from "../components/game/SecondaryDetailsModal";
import type { SecondaryCard } from "../types/game";
import { useStratagems } from "../hooks/queries/useFactionQueries";
import { useMissions, useDeploymentPatterns } from "../hooks/queries/useMissionQueries";
import { useGameEvents } from "../hooks/queries/useGamesQueries";
import { type RestGameEvent, normalizeWsEvent } from "../components/game/eventFormatting";
import { buildScoringHeatmapData } from "../components/game/vpUtils";
import { PlayerScoringHeatmap } from "../components/game/PlayerScoringHeatmap";
import { ScoringDetailModal, type CellSelection } from "../components/game/ScoringDetailModal";
import type { GameEvent, Phase } from "../types/game";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Spinner } from "@/components/ui/spinner";
import { Badge } from "@/components/ui/badge";
import { ShareSpectateButton } from "../components/game/ShareSpectateButton";

export function GamePage() {
  const { id: gameId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { gameState, events, opponentConnected, setEvents } = useGameStore();

  const token = getToken();

  const { connected, reconnecting, sendAction } = useGameConnection(gameId!, token);

  useEffect(() => {
    if (gameState?.gameId === gameId && gameState?.status === "setup") {
      void navigate(`/game/${gameId}/setup`);
    }
  }, [gameState?.gameId, gameState?.status, gameId, navigate]);

  useEffect(() => {
    return () => {
      useGameStore.getState().reset();
    };
  }, []);

  // Seed the event log with the persisted history so the log isn't empty for
  // players who refresh or join mid-game. Live WS events arriving in parallel
  // are deduped by id in the store.
  const { data: historicalEvents } = useGameEvents(gameId!);
  useEffect(() => {
    if (!historicalEvents) return;
    const seeded: GameEvent[] = (historicalEvents as RestGameEvent[]).map((e) => ({
      id: e.id,
      eventType: e.eventType,
      playerNumber: e.playerNumber ?? undefined,
      round: e.round ?? undefined,
      phase: (e.phase ?? undefined) as Phase | undefined,
      data: e.eventData ?? undefined,
      createdAt: e.createdAt,
    }));
    setEvents(seeded);
  }, [historicalEvents, setEvents]);

  const [showStratagems, setShowStratagems] = useState(false);
  const [showLog, setShowLog] = useState(false);
  const [showConcedeModal, setShowConcedeModal] = useState(false);
  const [showAbandonModal, setShowAbandonModal] = useState(false);
  const [showRevertModal, setShowRevertModal] = useState(false);
  const [showCPCapOverride, setShowCPCapOverride] = useState(false);
  const [opponentDetailsCard, setOpponentDetailsCard] = useState<SecondaryCard | null>(null);
  const [scoringSelection, setScoringSelection] = useState<CellSelection | null>(null);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;
  const isMyTurn = myPlayer?.playerNumber === gameState?.activePlayer;

  const {
    data: stratagems = [],
    isError: stratagemsError,
    refetch: refetchStratagems,
  } = useStratagems(myPlayer?.factionId);
  const { data: allMissions = [] } = useMissions();
  const { data: deploymentPatterns = [] } = useDeploymentPatterns();
  const deploymentPattern =
    deploymentPatterns.find((p) => p.id === gameState?.board?.deploymentPatternId) ?? null;

  // Mission is now per-player in 11th edition. Resolve it from the viewing
  // player's missionId against the mission list.
  const currentMission = allMissions.find((m) => m.id === myPlayer?.missionId) ?? null;

  const availableStratagems = stratagems.filter((s) => {
    if (!gameState) return false;

    const phase = gameState.currentPhase;
    const phases = s.phases ?? [];
    const phaseMatch =
      phases.length === 0 ||
      phases.some((p) => p === "Any phase" || p.toLowerCase().includes(phase.toLowerCase()));

    const turnMatch =
      !s.playerTurn ||
      s.playerTurn === "Either player's turn" ||
      (isMyTurn ? s.playerTurn === "Your turn" : s.playerTurn === "Opponent's turn");

    const detachmentMatch = !s.detachmentId || s.detachmentId === myPlayer?.detachmentId;

    const isChallenger = s.type?.startsWith("Challenger \u2013 ") ?? false;

    return phaseMatch && turnMatch && detachmentMatch && !isChallenger;
  });

  // 11e scoring fires automatically on the backend at end-of-command /
  // end-of-turn / end-of-battle: Layer-1 awards auto-apply and Layer-2 awards
  // raise a confirmation prompt. Advancing the phase no longer needs a manual
  // scoring reminder.
  const handleAdvancePhase = useCallback(() => {
    sendAction("advance_phase");
  }, [sendAction]);

  const handleAdjustCP = useCallback(
    (delta: number) => {
      if (delta > 0 && (myPlayer?.cpGainedThisRound ?? 0) >= 1) {
        setShowCPCapOverride(true);
        return;
      }
      sendAction("adjust_cp", { delta });
    },
    [sendAction, myPlayer?.cpGainedThisRound],
  );

  const handleConfirmCPCapOverride = useCallback(() => {
    sendAction("adjust_cp", { delta: 1, force: true });
    setShowCPCapOverride(false);
  }, [sendAction]);

  const handleAdjustVPManual = useCallback(
    (category: string, delta: number) => {
      sendAction("adjust_vp_manual", { category, delta });
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

  const handleDrawSecondaries = useCallback(() => {
    sendAction("draw_secondaries");
  }, [sendAction]);

  const handleSetObjectiveControl = useCallback(
    (objectiveIndex: number, player: number) => {
      sendAction("set_objective_control", { objectiveIndex, player });
    },
    [sendAction],
  );

  const handleConfirmAward = useCallback(
    (promptId: string, count: number) => {
      sendAction("confirm_award", { promptId, count });
    },
    [sendAction],
  );

  if (!gameState || !myPlayer) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="flex items-center gap-2">
          <Spinner />
          <span className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
            {connected ? "Loading game" : "Connecting"}
          </span>
        </div>
      </div>
    );
  }

  const opponentVP = opponent ? opponent.vpPrimary + opponent.vpSecondary + opponent.vpPaint : 0;

  const heatmapData = buildScoringHeatmapData(events.map(normalizeWsEvent), [myPlayer, opponent], {
    roundCount: Math.max(gameState.currentRound, 1),
  });
  const myStats = heatmapData.statsByPlayerNumber[myPlayer.playerNumber];
  const opponentStats = opponent
    ? (heatmapData.statsByPlayerNumber[opponent.playerNumber] ?? null)
    : null;

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
    <div className="relative flex h-screen flex-col overflow-hidden bg-background text-foreground">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            "linear-gradient(var(--primary) 1px, transparent 1px), linear-gradient(90deg, var(--primary) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      />

      {/* Turn Banner */}
      <div
        className={`relative z-10 border-b px-4 py-3 text-center font-mono text-sm uppercase tracking-widest backdrop-blur-sm ${
          isMyTurn
            ? "border-primary/50 bg-primary/15 text-primary shadow-[0_0_20px_var(--primary)]"
            : "border-border/60 bg-background/60 text-muted-foreground"
        }`}
      >
        Battle Round {gameState.currentRound} — {isMyTurn ? "Your" : `${opponent?.username}'s`} Turn
        — {PHASE_LABELS[gameState.currentPhase]} Phase
        <div className="absolute right-2 top-1/2 -translate-y-1/2">
          <ShareSpectateButton gameId={gameState.gameId} size="icon" variant="ghost" />
        </div>
      </div>

      {reconnecting && (
        <div
          role="status"
          aria-live="polite"
          className="relative z-10 flex items-center justify-center gap-2 border-b border-amber-500/40 bg-amber-500/10 px-4 py-2 font-mono text-[10px] uppercase tracking-widest text-amber-300"
        >
          <Spinner className="text-amber-300" />
          Reconnecting to server...
        </div>
      )}

      {/* Round & Phase */}
      <div className="relative z-10 space-y-3 border-b border-border/60 bg-background/40 px-4 py-3 backdrop-blur-sm">
        <RoundIndicator
          currentRound={gameState.currentRound}
          currentTurn={gameState.currentTurn}
          maxRounds={5}
        />
        <PhaseTracker currentPhase={gameState.currentPhase} phases={PHASE_ORDER} />
      </div>

      {/* Main Content */}
      <div className="relative z-0 min-h-0 flex-1 overflow-auto px-4 py-4">
        <div className="mx-auto max-w-3xl space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-3">
              <HUDFrame label={`${myPlayer.username} — ${myPlayer.factionName}`}>
                <div className="space-y-3 py-1">
                  <div className="flex items-center justify-center">
                    <PlayerAvatar
                      avatarUrl={myPlayer.avatarUrl}
                      username={myPlayer.username}
                      size="md"
                    />
                  </div>
                  <CPCounter
                    cp={myPlayer.cp}
                    canGainCP={myPlayer.cpGainedThisRound < 1}
                    onAdjust={handleAdjustCP}
                  />
                  <VPCounter
                    vpPrimary={myPlayer.vpPrimary}
                    vpSecondary={myPlayer.vpSecondary}
                    vpPaint={myPlayer.vpPaint}
                    onAdjust={handleAdjustVPManual}
                  />
                </div>
              </HUDFrame>
              <PlayerScoringHeatmap
                username={myPlayer.username}
                stats={myStats}
                rounds={heatmapData.rounds}
                intensityMax={heatmapData.intensityMax}
                onCellClick={(round, category) =>
                  setScoringSelection({
                    playerNumber: myPlayer.playerNumber,
                    username: myPlayer.username,
                    round,
                    category,
                  })
                }
              />
            </div>

            {opponent && (
              <div className="flex flex-col gap-3">
                <HUDFrame label={`${opponent.username} — ${opponent.factionName}`}>
                  <div className="space-y-3 py-1">
                    <div className="flex items-center justify-center">
                      <PlayerAvatar
                        avatarUrl={opponent.avatarUrl}
                        username={opponent.username}
                        size="md"
                      />
                    </div>
                    <div className="text-center">
                      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                        Command Points
                      </p>
                      <span className="mt-1 block font-mono text-3xl font-bold text-primary tabular-nums">
                        {opponent.cp}
                      </span>
                    </div>
                    <div className="text-center">
                      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                        Victory Points
                      </p>
                      <span className="mt-1 block font-mono text-3xl font-bold text-primary tabular-nums">
                        {opponentVP}
                      </span>
                    </div>
                    {!opponentConnected && (
                      <div className="flex justify-center">
                        <Badge
                          variant="outline"
                          role="status"
                          aria-label="Opponent disconnected"
                          className="border-amber-500/50 font-mono text-[10px] uppercase tracking-widest text-amber-300"
                        >
                          Disconnected
                        </Badge>
                      </div>
                    )}
                  </div>
                </HUDFrame>
                {opponentStats && (
                  <PlayerScoringHeatmap
                    username={opponent.username}
                    stats={opponentStats}
                    rounds={heatmapData.rounds}
                    intensityMax={heatmapData.intensityMax}
                    onCellClick={(round, category) =>
                      setScoringSelection({
                        playerNumber: opponent.playerNumber,
                        username: opponent.username,
                        round,
                        category,
                      })
                    }
                  />
                )}
              </div>
            )}
          </div>

          {/* Layer-2 scoring confirmations awaiting this player */}
          <ScorePromptsPanel
            prompts={myPlayer.pendingScorePrompts ?? []}
            onConfirm={handleConfirmAward}
          />

          {/* Battlefield objective control */}
          <BoardView
            board={gameState.board}
            pattern={deploymentPattern}
            onSetControl={handleSetObjectiveControl}
          />

          {/* Opponent's Active Secondaries */}
          {opponent && (opponent.secondaryHand ?? []).length > 0 && (
            <div className="rounded-sm border border-border/40 bg-background/40 p-3">
              <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                {`${opponent.username}'s Active Secondaries (${opponent.secondaryMode === "tactical" ? "Tactical" : "Fixed"})`}
              </h3>
              <div className="mt-2 space-y-2">
                {(opponent.secondaryHand ?? []).map((s) => (
                  <button
                    type="button"
                    key={s.id}
                    onClick={() => setOpponentDetailsCard(s)}
                    title="View full details"
                    className="block w-full cursor-pointer rounded-sm border border-border/60 bg-background/40 p-2 text-left transition-colors hover:border-primary/50"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <span className="text-sm font-medium text-foreground">{s.name}</span>
                    </div>
                    <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{s.text}</p>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Secondary Missions */}
          <SecondaryPanel
            mode={myPlayer.secondaryMode}
            secondaryHand={myPlayer.secondaryHand ?? []}
            secondaryScored={myPlayer.secondaryScored ?? []}
            secondaryDeck={myPlayer.secondaryDeck ?? []}
            currentPhase={gameState.currentPhase}
            isMyTurn={isMyTurn}
            onDraw={handleDrawSecondaries}
          />

          {/* Mission Info */}
          <MissionInfo mission={currentMission} />

          {/* Stratagem Panel */}
          <section className="space-y-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowStratagems(!showStratagems)}
              className="w-full justify-between font-mono uppercase tracking-widest"
              disabled={stratagemsError}
            >
              <span className="flex items-center gap-2">
                <Zap className="size-4" />
                {stratagemsError
                  ? "Stratagems unavailable"
                  : `Stratagems (${availableStratagems.length} available)`}
              </span>
              {showStratagems ? (
                <ChevronUp className="size-4" />
              ) : (
                <ChevronDown className="size-4" />
              )}
            </Button>
            {stratagemsError && (
              <div
                role="alert"
                className="flex items-center justify-between gap-2 rounded-sm border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-200"
              >
                <span>Stratagems failed to load.</span>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() => void refetchStratagems()}
                  className="font-mono uppercase tracking-widest"
                >
                  Retry
                </Button>
              </div>
            )}
            {showStratagems && !stratagemsError && (
              <StratagemPanel
                stratagems={availableStratagems}
                currentCP={myPlayer.cp}
                usedThisPhase={myPlayer.stratagemsUsedThisPhase ?? []}
                onUse={handleUseStratagem}
              />
            )}
          </section>

          {/* Game Log */}
          <section className="space-y-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowLog(!showLog)}
              className="w-full justify-between font-mono uppercase tracking-widest"
            >
              <span className="flex items-center gap-2">
                <ScrollText className="size-4" />
                Game Log
              </span>
              {showLog ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
            </Button>
            {showLog && <GameLog events={events} players={gameState?.players} />}
          </section>
        </div>
      </div>

      {/* Bottom Action Bar */}
      <div className="relative z-10 flex flex-wrap gap-2 border-t border-border/60 bg-background/60 p-3 backdrop-blur-sm">
        {isMyTurn && (
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowRevertModal(true)}
              title="Step back one phase"
              className="font-mono uppercase tracking-widest"
            >
              ← Revert
            </Button>
            <Button
              type="button"
              onClick={handleAdvancePhase}
              className="flex-1 gap-2 font-mono uppercase tracking-widest"
            >
              <Forward className="size-4" />
              Advance Phase
            </Button>
          </>
        )}
        <Button
          type="button"
          variant="destructive"
          onClick={() => setShowConcedeModal(true)}
          className="gap-1 font-mono uppercase tracking-widest"
        >
          <Flag className="size-4" />
          Concede
        </Button>
        <Button
          type="button"
          variant="outline"
          onClick={() => setShowAbandonModal(true)}
          className="gap-1 font-mono uppercase tracking-widest"
        >
          <Handshake className="size-4" />
          Abandon
        </Button>
      </div>

      {/* Revert Phase Confirmation */}
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

      {/* Concede Confirmation */}
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

      {/* Abandon Request */}
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

      {/* Abandon Request Received */}
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

      {showCPCapOverride && (
        <ConfirmModal
          title="CP Gain Cap Reached"
          message="You have already gained your bonus CP this battle round. Increase CP beyond the cap?"
          confirmLabel="Increase CP"
          cancelLabel="Cancel"
          variant="default"
          onConfirm={handleConfirmCPCapOverride}
          onCancel={() => setShowCPCapOverride(false)}
        />
      )}

      <SecondaryDetailsModal
        secondary={opponentDetailsCard}
        onClose={() => setOpponentDetailsCard(null)}
      />

      <ScoringDetailModal
        selection={scoringSelection}
        onClose={() => setScoringSelection(null)}
        events={heatmapData.normalizedEvents}
      />
    </div>
  );
}
