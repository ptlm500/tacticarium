import { useCallback, useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  ChevronDown,
  ChevronUp,
  Flag,
  Forward,
  Handshake,
  ScrollText,
  Sparkles,
  Zap,
} from "lucide-react";
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
import { useGameEvents } from "../hooks/queries/useGamesQueries";
import { type RestGameEvent } from "../components/game/eventFormatting";
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

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;
  const isMyTurn = myPlayer?.playerNumber === gameState?.activePlayer;

  const [scoringPromptItems, setScoringPromptItems] = useState<ScoringPromptItem[] | null>(null);
  const [opponentTurnPromptItems, setOpponentTurnPromptItems] = useState<
    ScoringPromptItem[] | null
  >(null);
  const [showDrawPrompt, setShowDrawPrompt] = useState(false);
  // Tracks the previous (phase, activePlayer, status) to detect the moment the
  // opponent's Fight phase ends and pop a reactive prompt for end-of-opponent-turn
  // secondaries. Seeded on first observation so a page reload mid-turn doesn't
  // retroactively trigger.
  const prevTurnState = useRef<{
    phase: Phase;
    activePlayer: number;
    status: string;
  } | null>(null);

  useEffect(() => {
    if (!gameState || !myPlayer) return;
    const prev = prevTurnState.current;
    const opponentNum = myPlayer.playerNumber === 1 ? 2 : 1;

    if (prev) {
      const prevWasOpponentFight =
        prev.phase === "fight" && prev.activePlayer === opponentNum && prev.status === "active";
      const stillOpponentFight =
        gameState.currentPhase === "fight" && gameState.activePlayer === opponentNum;

      if (prevWasOpponentFight && !stillOpponentFight) {
        const opponentTurnSecondaries = (myPlayer.activeSecondaries ?? []).filter(
          (s) => s.scoringTiming === "end_of_opponent_turn",
        );
        if (opponentTurnSecondaries.length > 0) {
          setOpponentTurnPromptItems(
            myPlayer.secondaryMode === "fixed"
              ? [
                  {
                    kind: "fixed_secondary",
                    secondaries: opponentTurnSecondaries,
                    timing: "end_of_opponent_turn",
                  },
                ]
              : [{ kind: "secondary", timing: "end_of_opponent_turn" }],
          );
        }
      }
    }

    prevTurnState.current = {
      phase: gameState.currentPhase,
      activePlayer: gameState.activePlayer,
      status: gameState.status,
    };
  }, [gameState, myPlayer]);

  const {
    data: stratagems = [],
    isError: stratagemsError,
    refetch: refetchStratagems,
  } = useStratagems(myPlayer?.factionId);
  const { data: allMissions = [] } = useMissions(gameState?.missionPackId);
  const { data: allRules = [] } = useMissionRules(gameState?.missionPackId);

  const currentMission = allMissions.find((m) => m.id === gameState?.missionId) ?? null;
  const currentTwist = allRules.find((r) => r.id === gameState?.twistId) ?? null;

  const availableStratagems = stratagems.filter((s) => {
    if (!gameState) return false;

    const phase = gameState.currentPhase;
    const phaseMatch =
      s.phase === "Any phase" || s.phase.toLowerCase().includes(phase.toLowerCase());

    const turnMatch = isMyTurn
      ? s.turn === "Your turn" || s.turn === "Either player's turn"
      : s.turn === "Opponent's turn" || s.turn === "Either player's turn";

    const detachmentMatch = !s.detachmentId || s.detachmentId === myPlayer?.detachmentId;

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

    if (currentMission) {
      if (scoringTiming === "end_of_command_phase") {
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
        if (isFightPhase && isSecondPlayerTurn) {
          items.push({
            kind: "end_of_round_primary",
            missionName: currentMission.name,
            note: "Both players score at the end of each battle round. Make sure your opponent has scored too.",
          });
        }
      }

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

    if (isFightPhase) {
      if (myPlayer.secondaryMode === "fixed") {
        const fixedSecondaries = (myPlayer.activeSecondaries ?? []).filter(
          (s) => s.isFixed && (s.scoringTiming ?? "end_of_own_turn") === "end_of_own_turn",
        );
        if (fixedSecondaries.length > 0) {
          items.push({
            kind: "fixed_secondary",
            secondaries: fixedSecondaries,
            timing: "end_of_own_turn",
          });
        }
      } else {
        const hasOwnTurnSecondary = (myPlayer.activeSecondaries ?? []).some(
          (s) => (s.scoringTiming ?? "end_of_own_turn") === "end_of_own_turn",
        );
        if (hasOwnTurnSecondary) {
          items.push({ kind: "secondary", timing: "end_of_own_turn" });
        }
      }
    }

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

  const handleScoreVP = useCallback(
    (
      category: string,
      delta: number,
      scoringSlot?: PrimaryScoringSlot,
      scoringRuleLabel?: string,
    ) => {
      const data: Record<string, unknown> = { category, delta };
      if (scoringSlot) data.scoringSlot = scoringSlot;
      if (scoringRuleLabel) data.scoringRuleLabel = scoringRuleLabel;
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
    <div className="relative flex min-h-screen flex-col overflow-hidden bg-background text-foreground">
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
      <div className="relative z-0 flex-1 overflow-auto px-4 py-4">
        <div className="mx-auto max-w-3xl space-y-4">
          {/* Your State */}
          <HUDFrame label={`${myPlayer.username} — ${myPlayer.factionName}`}>
            <div className="space-y-3 py-1">
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
              {currentMission &&
                currentMission.scoringRules &&
                currentMission.scoringRules.length > 0 && (
                  <MissionScoring
                    scoringRules={currentMission.scoringRules ?? []}
                    currentRound={gameState.currentRound}
                    missionScoringTiming={currentMission.scoringTiming ?? "end_of_command_phase"}
                    onScore={(vp, slot, label) => handleScoreVP("primary", vp, slot, label)}
                  />
                )}
              <PrimaryScoreHistory
                scoredSlots={myPlayer.vpPrimaryScoredSlots ?? {}}
                onUndo={handleUndoPrimaryScore}
              />
            </div>
          </HUDFrame>

          {/* Opponent State */}
          {opponent && (
            <div className="rounded-sm border border-border/40 bg-background/40 p-3">
              <div className="flex items-center justify-between gap-2">
                <h2 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                  {opponent.username} — {opponent.factionName}
                </h2>
                {!opponentConnected && (
                  <Badge
                    variant="outline"
                    role="status"
                    aria-label="Opponent disconnected"
                    className="border-amber-500/50 font-mono text-[10px] uppercase tracking-widest text-amber-300"
                  >
                    Disconnected
                  </Badge>
                )}
              </div>
              <div className="mt-2 flex gap-6 font-mono text-sm tabular-nums">
                <span>
                  <span className="text-muted-foreground">CP:</span> {opponent.cp}
                </span>
                <span>
                  <span className="text-muted-foreground">VP:</span> {opponentVP}
                </span>
              </div>
              {(opponent.activeSecondaries ?? []).length > 0 && (
                <div className="mt-3 space-y-2">
                  <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                    Active Secondaries (
                    {opponent.secondaryMode === "tactical" ? "Tactical" : "Fixed"})
                  </h3>
                  {(opponent.activeSecondaries ?? []).map((s) => (
                    <div
                      key={s.id}
                      className="rounded-sm border border-border/60 bg-background/40 p-2"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <span className="text-sm font-medium text-foreground">{s.name}</span>
                        <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                          {s.maxVp} VP max
                        </span>
                      </div>
                      <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                        {s.description}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
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

          {/* Challenger Card Banner */}
          {opponent &&
            gameState.currentPhase === "command" &&
            totalVP + 6 <= opponentVP &&
            !myPlayer.isChallenger && (
              <div className="rounded-sm border border-amber-500/40 bg-amber-500/10 p-3 text-center">
                <p className="text-xs text-amber-200">
                  You are trailing by{" "}
                  <Badge variant="outline" className="border-amber-400/60 font-mono text-amber-300">
                    {opponentVP - totalVP} VP
                  </Badge>{" "}
                  — eligible for a Challenger Card!
                </p>
                <Button
                  type="button"
                  size="sm"
                  onClick={handleDrawChallengerCard}
                  className="mt-2 gap-1 bg-amber-600 text-white hover:bg-amber-700"
                >
                  <Sparkles className="size-3" />
                  Draw Challenger Card
                </Button>
              </div>
            )}

          {/* Active Challenger Card */}
          {myPlayer.isChallenger && myPlayer.challengerCardId && (
            <div className="rounded-sm border border-purple-500/40 bg-purple-500/10 p-3">
              <p className="font-mono text-xs uppercase tracking-widest text-purple-300">
                Active Challenger Card
              </p>
              <Button
                type="button"
                size="sm"
                onClick={handleScoreChallenger}
                className="mt-2 bg-purple-600 text-white hover:bg-purple-700"
              >
                Complete Mission (+3 VP)
              </Button>
            </div>
          )}

          {/* Mission Info */}
          <MissionInfo mission={currentMission} twist={currentTwist} />

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
            {showLog && <GameLog events={events} />}
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

      {/* Scoring Prompt */}
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

      {/* Reactive prompt fired when opponent's Fight phase ends — for
          secondaries that score at the end of the opponent's turn (e.g. Sabotage). */}
      {opponentTurnPromptItems && (
        <ScoringPrompt
          items={opponentTurnPromptItems}
          onScore={handleScoreVP}
          activeSecondaries={myPlayer.activeSecondaries ?? []}
          onAchieveSecondary={handleAchieveSecondary}
          onDiscardSecondary={handleDiscardSecondary}
          canGainCP={myPlayer.cpGainedThisRound < 1}
          onScoreFixedVP={(delta) => handleScoreVP("secondary", delta)}
          onConfirm={() => setOpponentTurnPromptItems(null)}
          onCancel={() => setOpponentTurnPromptItems(null)}
          title="Opponent's Turn Ended"
          description="Score any secondaries that resolve at the end of your opponent's turn."
          confirmLabel="Done"
          cancelLabel="Dismiss"
        />
      )}

      {/* Draw Prompt */}
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
