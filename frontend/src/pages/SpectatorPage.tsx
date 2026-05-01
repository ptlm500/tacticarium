import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { ChevronDown, ChevronUp, Eye, ScrollText } from "lucide-react";
import { useGameStore } from "../stores/gameStore";
import { useSpectatorConnection } from "../hooks/useSpectatorConnection";
import { useGameEvents } from "../hooks/queries/useGamesQueries";
import { useMissions, useMissionRules } from "../hooks/queries/useMissionQueries";
import { PHASE_LABELS, PHASE_ORDER, type GameEvent, type Phase } from "../types/game";
import { type RestGameEvent } from "../components/game/eventFormatting";
import { PhaseTracker } from "../components/game/PhaseTracker";
import { RoundIndicator } from "../components/game/RoundIndicator";
import { MissionInfo } from "../components/game/MissionInfo";
import { GameLog } from "../components/game/GameLog";
import { SpectatorPlayerPanel } from "../components/game/SpectatorPlayerPanel";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Spinner } from "@/components/ui/spinner";
import { Badge } from "@/components/ui/badge";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";

export function SpectatorPage() {
  const { id: gameId } = useParams<{ id: string }>();
  const { gameState, events, setEvents } = useGameStore();

  const { connected, reconnecting } = useSpectatorConnection(gameId!);

  useEffect(() => {
    return () => {
      useGameStore.getState().reset();
    };
  }, []);

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

  const { data: allMissions = [] } = useMissions(gameState?.missionPackId);
  const { data: allRules = [] } = useMissionRules(gameState?.missionPackId);
  const currentMission = allMissions.find((m) => m.id === gameState?.missionId) ?? null;
  const currentTwist = allRules.find((r) => r.id === gameState?.twistId) ?? null;

  const [showLog, setShowLog] = useState(false);

  if (!gameState) {
    return (
      <SpectatorShell>
        <div className="flex min-h-[60vh] items-center justify-center">
          <div className="flex items-center gap-2">
            <Spinner />
            <span className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
              {connected ? "Loading game" : "Connecting"}
            </span>
          </div>
        </div>
      </SpectatorShell>
    );
  }

  if (gameState.status === "setup") {
    return (
      <SpectatorShell>
        <main className="relative z-0 mx-auto flex min-h-[60vh] max-w-xl items-center px-4 py-12">
          <HUDFrame label="Game Not Started">
            <div className="space-y-3 py-6 text-center">
              <Eye className="mx-auto size-8 text-muted-foreground" />
              <p className="font-mono text-sm uppercase tracking-widest text-foreground">
                Awaiting Battle
              </p>
              <p className="text-sm text-muted-foreground">
                Players are still configuring this game. Check back once the battle begins.
              </p>
            </div>
          </HUDFrame>
        </main>
      </SpectatorShell>
    );
  }

  if (gameState.status === "completed" || gameState.status === "abandoned") {
    return (
      <SpectatorShell>
        <main className="relative z-0 mx-auto flex min-h-[60vh] max-w-xl items-center px-4 py-12">
          <HUDFrame label="Game Ended">
            <div className="space-y-3 py-6 text-center">
              <Eye className="mx-auto size-8 text-muted-foreground" />
              <p className="font-mono text-sm uppercase tracking-widest text-foreground">
                {gameState.status === "abandoned" ? "Mutually Abandoned" : "Battle Complete"}
              </p>
              <p className="text-sm text-muted-foreground">
                Spectator mode is only available for active games.
              </p>
            </div>
          </HUDFrame>
        </main>
      </SpectatorShell>
    );
  }

  const [player1, player2] = gameState.players;

  return (
    <SpectatorShell>
      <div
        className={`relative z-10 border-b px-4 py-3 text-center font-mono text-sm uppercase tracking-widest backdrop-blur-sm border-border/60 bg-background/60 text-muted-foreground`}
      >
        Battle Round {gameState.currentRound} —{" "}
        {gameState.activePlayer === 1
          ? player1?.username
          : (player2?.username ?? "Player " + gameState.activePlayer)}
        ’s Turn — {PHASE_LABELS[gameState.currentPhase]} Phase
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

      <div className="relative z-10 space-y-3 border-b border-border/60 bg-background/40 px-4 py-3 backdrop-blur-sm">
        <RoundIndicator
          currentRound={gameState.currentRound}
          currentTurn={gameState.currentTurn}
          maxRounds={5}
        />
        <PhaseTracker currentPhase={gameState.currentPhase} phases={PHASE_ORDER} />
      </div>

      <main className="relative z-0 flex-1 overflow-auto px-4 py-4">
        <div className="mx-auto max-w-7xl space-y-4">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {player1 && (
              <SpectatorPlayerPanel player={player1} isActive={gameState.activePlayer === 1} />
            )}
            {player2 && (
              <SpectatorPlayerPanel player={player2} isActive={gameState.activePlayer === 2} />
            )}
          </div>

          <MissionInfo mission={currentMission} twist={currentTwist} />

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
            {showLog && <GameLog events={events} players={gameState.players} />}
          </section>
        </div>
      </main>
    </SpectatorShell>
  );
}

function SpectatorShell({ children }: { children: React.ReactNode }) {
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
      <header className="relative z-10 flex items-center justify-between border-b border-border/50 bg-background/60 px-6 py-3 backdrop-blur-sm">
        <div className="flex items-baseline gap-3">
          <span className="font-mono text-base uppercase tracking-[0.3em] text-primary">
            Tacticarium
          </span>
          <Badge variant="outline" className="font-mono text-[10px] uppercase tracking-widest">
            <Eye className="mr-1 size-3" />
            Spectator
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          <ThemeSwitcher />
          <Button
            asChild
            variant="ghost"
            size="sm"
            className="font-mono text-[10px] uppercase tracking-widest"
          >
            <Link to="/">Home</Link>
          </Button>
        </div>
      </header>
      {children}
    </div>
  );
}
