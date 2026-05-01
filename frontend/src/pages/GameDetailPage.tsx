import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, Handshake, Skull, Swords, Trophy } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import {
  type RestGameEvent,
  buildPlayerStats,
  getEndReason,
  getRoundsPlayed,
} from "../components/game/vpUtils";
import { VPBreakdownTable } from "../components/game/VPBreakdownTable";
import { VPProgressionChart } from "../components/game/VPProgressionChart";
import { EventTimeline } from "../components/game/EventTimeline";
import { buildPlayerInfo, normalizeRestEvent } from "../components/game/eventFormatting";
import { useGame, useGameEvents } from "../hooks/queries/useGamesQueries";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Spinner } from "@/components/ui/spinner";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";

export function GameDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const { data: gameState, isPending: gameLoading } = useGame(id!);
  const { data: events, isPending: eventsLoading } = useGameEvents(id!);

  if (gameLoading || eventsLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="flex items-center gap-2">
          <Spinner />
          <span className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
            Loading
          </span>
        </div>
      </div>
    );
  }

  if (!gameState || !events) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="space-y-4 text-center">
          <p className="font-mono text-sm uppercase tracking-widest text-destructive">
            Game not found
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate("/history")}
            className="gap-1.5 font-mono text-[10px] uppercase tracking-widest"
          >
            <ArrowLeft className="size-3.5" />
            Back to History
          </Button>
        </div>
      </div>
    );
  }

  const myPlayer = gameState.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState.players.find((p) => p?.userId !== user?.id) ?? null;

  if (!myPlayer) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
          You were not a player in this game.
        </p>
      </div>
    );
  }

  const typedEvents = events as RestGameEvent[];
  const endReason = getEndReason(typedEvents);
  const roundsPlayed = getRoundsPlayed(typedEvents);
  const rounds = Array.from({ length: roundsPlayed }, (_, i) => i + 1);

  const myStats = buildPlayerStats(typedEvents, myPlayer.playerNumber, myPlayer.vpPaint);
  const opponentStats = opponent
    ? buildPlayerStats(typedEvents, opponent.playerNumber, opponent.vpPaint)
    : null;

  const isAbandoned = gameState.status === "abandoned";
  const isWinner = gameState.winnerId === user?.id;

  let heading: string;
  let HeadingIcon: typeof Trophy;
  let headingColor: string;
  if (isAbandoned) {
    heading = "Game Abandoned";
    HeadingIcon = Handshake;
    headingColor = "text-muted-foreground";
  } else if (!gameState.winnerId) {
    heading = "Draw";
    HeadingIcon = Swords;
    headingColor = "text-amber-400";
  } else if (isWinner) {
    heading = "Victory!";
    HeadingIcon = Trophy;
    headingColor = "text-emerald-400";
  } else {
    heading = "Defeat";
    HeadingIcon = Skull;
    headingColor = "text-destructive";
  }

  const reasonLabel =
    endReason === "concede"
      ? "Ended by concession"
      : endReason === "abandoned"
        ? "Mutually abandoned"
        : endReason === "rounds_complete"
          ? `Completed after ${roundsPlayed} rounds`
          : null;

  const normalizedEvents = typedEvents.map(normalizeRestEvent);
  const timelinePlayers = buildPlayerInfo(gameState.players);

  return (
    <div className="relative min-h-screen overflow-hidden bg-background text-foreground">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            "linear-gradient(var(--primary) 1px, transparent 1px), linear-gradient(90deg, var(--primary) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_0%,var(--background)_85%)]"
      />

      <header className="relative z-10 flex items-center justify-between border-b border-border/50 bg-background/60 px-6 py-3 backdrop-blur-sm">
        <div className="flex items-baseline gap-3">
          <span className="font-mono text-base uppercase tracking-[0.3em] text-primary">
            Tacticarium
          </span>
          <span className="hidden font-mono text-[10px] uppercase tracking-widest text-muted-foreground/70 sm:inline">
            Battle Report
          </span>
        </div>
        <div className="flex items-center gap-2">
          <ThemeSwitcher />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate("/history")}
            className="gap-1.5 font-mono text-[10px] uppercase tracking-widest"
          >
            <ArrowLeft className="size-3.5" />
            Back
          </Button>
        </div>
      </header>

      <main className="relative z-0 mx-auto max-w-3xl space-y-6 px-4 py-8">
        <HUDFrame label="Outcome">
          <div className="py-1 text-center">
            <HeadingIcon className={`mx-auto size-8 ${headingColor}`} />
            <h2 className={`mt-2 font-mono text-2xl uppercase tracking-[0.3em] ${headingColor}`}>
              {heading}
            </h2>
            {reasonLabel && (
              <p className="mt-2 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                {reasonLabel}
              </p>
            )}
            {gameState.missionName && (
              <p className="mt-2 font-mono text-xs text-foreground/80">{gameState.missionName}</p>
            )}
            <p className="mt-1 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              {myPlayer.factionName} vs {opponent?.factionName ?? "Unknown"}
            </p>
          </div>
        </HUDFrame>

        <HUDFrame label="VP Breakdown">
          <div className="py-1">
            <VPBreakdownTable
              myStats={myStats}
              opponentStats={opponentStats}
              myUsername={myPlayer.username}
              opponentUsername={opponent?.username ?? null}
              rounds={rounds}
            />
          </div>
        </HUDFrame>

        {rounds.length > 0 && (
          <HUDFrame label="VP Progression">
            <div className="py-1">
              <VPProgressionChart
                myStats={myStats}
                opponentStats={opponentStats}
                myUsername={myPlayer.username}
                opponentUsername={opponent?.username ?? null}
                rounds={rounds}
              />
            </div>
          </HUDFrame>
        )}

        <div className="flex flex-wrap gap-x-6 gap-y-2 font-mono text-xs text-muted-foreground">
          <div>
            Rounds played: <span className="text-foreground tabular-nums">{roundsPlayed}</span>
          </div>
          <div>
            {myPlayer.username} stratagems:{" "}
            <span className="text-foreground tabular-nums">{myStats.stratagemsUsed}</span>
          </div>
          {opponentStats && (
            <div>
              {opponent!.username} stratagems:{" "}
              <span className="text-foreground tabular-nums">{opponentStats.stratagemsUsed}</span>
            </div>
          )}
        </div>

        <HUDFrame label="Event Timeline">
          <div className="py-1">
            <EventTimeline
              events={normalizedEvents}
              defaultFilter="highlights"
              players={timelinePlayers}
            />
          </div>
        </HUDFrame>
      </main>
    </div>
  );
}
