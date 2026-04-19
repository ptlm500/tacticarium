import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ArrowLeft, Skull, Swords, Handshake, Trophy } from "lucide-react";
import { GameState, PlayerState } from "../../types/game";
import { gamesApi } from "../../api/games";
import { type RestGameEvent, buildPlayerStats, getEndReason, getRoundsPlayed } from "./vpUtils";
import { VPBreakdownTable } from "./VPBreakdownTable";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Spinner } from "@/components/ui/spinner";
import { Separator } from "@/components/ui/separator";

interface Props {
  gameState: GameState;
  myPlayer: PlayerState;
  opponent: PlayerState | null;
  currentUserId: string;
}

export function GameSummary({ gameState, myPlayer, opponent, currentUserId }: Props) {
  const [events, setEvents] = useState<RestGameEvent[] | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    gamesApi
      .getEvents(gameState.gameId)
      .then((data) => setEvents(data as RestGameEvent[]))
      .catch(() => setEvents([]))
      .finally(() => setLoading(false));
  }, [gameState.gameId]);

  const isWinner = gameState.winnerId === currentUserId;
  const isAbandoned = gameState.status === "abandoned";

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

  if (loading || !events) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="flex items-center gap-2">
          <Spinner />
          <span className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
            Loading summary
          </span>
        </div>
      </div>
    );
  }

  const endReason = getEndReason(events);
  const roundsPlayed = getRoundsPlayed(events);
  const rounds = Array.from({ length: roundsPlayed }, (_, i) => i + 1);

  const myStats = buildPlayerStats(events, myPlayer.playerNumber, myPlayer.vpPaint);
  const opponentStats = opponent
    ? buildPlayerStats(events, opponent.playerNumber, opponent.vpPaint)
    : null;

  const reasonLabel =
    endReason === "concede"
      ? "Ended by concession"
      : endReason === "abandoned"
        ? "Mutually abandoned"
        : endReason === "rounds_complete"
          ? `Completed after ${roundsPlayed} rounds`
          : null;

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

      <main className="relative z-0 mx-auto max-w-3xl px-4 py-10">
        <HUDFrame label="Battle Report">
          <div className="space-y-4 py-2">
            <div className="text-center">
              <HeadingIcon className={`mx-auto size-10 ${headingColor}`} />
              <h1 className={`mt-2 font-mono text-3xl uppercase tracking-[0.3em] ${headingColor}`}>
                {heading}
              </h1>
              {reasonLabel && (
                <p className="mt-2 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                  {reasonLabel}
                </p>
              )}
            </div>

            <Separator />

            <VPBreakdownTable
              myStats={myStats}
              opponentStats={opponentStats}
              myUsername={myPlayer.username}
              opponentUsername={opponent?.username ?? null}
              rounds={rounds}
            />

            <Separator />

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
                  <span className="text-foreground tabular-nums">
                    {opponentStats.stratagemsUsed}
                  </span>
                </div>
              )}
            </div>

            <div className="flex justify-center pt-2">
              <Button asChild className="gap-2 font-mono uppercase tracking-widest">
                <Link to="/">
                  <ArrowLeft className="size-4" />
                  Back to Lobby
                </Link>
              </Button>
            </div>
          </div>
        </HUDFrame>
      </main>
    </div>
  );
}
