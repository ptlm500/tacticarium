import { useEffect, useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Check, Copy, Swords } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { useGameStore } from "../stores/gameStore";
import { useGameConnection } from "../hooks/useGameState";
import { getToken } from "../api/client";
import { Faction, Detachment } from "../types/faction";
import { ForceDisposition } from "../types/mission";
import { FactionPicker } from "../components/setup/FactionPicker";
import { DetachmentPicker } from "../components/setup/DetachmentPicker";
import { SidePicker } from "../components/setup/SidePicker";
import { ForceDispositionPicker } from "../components/setup/ForceDispositionPicker";
import { FirstPlayerPicker } from "../components/setup/FirstPlayerPicker";
import { SecondaryModePicker } from "../components/setup/SecondaryModePicker";
import { FixedSecondaryPicker } from "../components/setup/FixedSecondaryPicker";
import { ArmyPaintedToggle } from "../components/setup/ArmyPaintedToggle";
import { useFactions, useDetachments } from "../hooks/queries/useFactionQueries";
import { useForceDispositions, useSecondaryCards } from "../hooks/queries/useMissionQueries";

// Mirrors the backend FixedSecondaryCount: a fixed-mode player keeps this many
// secondary cards for the whole game.
const FIXED_SECONDARY_COUNT = 2;
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Spinner } from "@/components/ui/spinner";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";
import { ShareSpectateButton } from "../components/game/ShareSpectateButton";
import { cn } from "@/lib/utils";

export function GameSetupPage() {
  const { id: gameId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { gameState } = useGameStore();

  const token = getToken();

  const { connected, reconnecting, sendAction } = useGameConnection(gameId!, token);

  const { data: factions = [] } = useFactions();
  const { data: dispositions = [] } = useForceDispositions();
  const { data: secondaryCards = [] } = useSecondaryCards();

  const [copied, setCopied] = useState(false);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;

  const { data: detachments = [] } = useDetachments(myPlayer?.factionId);

  // 11e: a player picks their force disposition from the set granted by their
  // chosen detachment. (Multi-detachment / DP-budget selection isn't modelled
  // yet — the engine holds a single detachment per player.) Fall back to the
  // full list when the detachment doesn't enumerate dispositions.
  const selectedDetachment = detachments.find((d) => d.id === myPlayer?.detachmentId) ?? null;
  const allowedDispositionIds = selectedDetachment?.forceDispositions ?? [];
  const availableDispositions =
    allowedDispositionIds.length > 0
      ? dispositions.filter((d) => allowedDispositionIds.includes(d.id))
      : dispositions;

  useEffect(() => {
    if (gameState?.gameId === gameId && gameState?.status === "active") {
      void navigate(`/game/${gameId}`);
    }
  }, [gameState?.gameId, gameState?.status, gameId, navigate]);

  useEffect(() => {
    return () => {
      useGameStore.getState().reset();
    };
  }, []);

  const handleSelectFaction = useCallback(
    (faction: Faction) => {
      sendAction("select_faction", {
        factionId: faction.id,
        factionName: faction.name,
      });
    },
    [sendAction],
  );

  const handleSelectDetachment = useCallback(
    (detachment: Detachment) => {
      sendAction("select_detachment", {
        detachmentId: detachment.id,
        detachmentName: detachment.name,
      });
    },
    [sendAction],
  );

  const handleSelectSide = useCallback(
    (side: "attacker" | "defender") => {
      sendAction("select_side", { side });
    },
    [sendAction],
  );

  const handleSelectDisposition = useCallback(
    (disposition: ForceDisposition) => {
      sendAction("select_force_disposition", {
        disposition: disposition.id,
        dispositionName: disposition.name,
      });
    },
    [sendAction],
  );

  const handleSelectFirstPlayer = useCallback(
    (playerNumber: 1 | 2) => {
      sendAction("select_first_turn_player", { firstTurnPlayer: playerNumber });
    },
    [sendAction],
  );

  const handleRandomFirstPlayer = useCallback(() => {
    const playerNumber = Math.random() < 0.5 ? 1 : 2;
    sendAction("select_first_turn_player", { firstTurnPlayer: playerNumber });
  }, [sendAction]);

  const handleModeChange = useCallback(
    (mode: "fixed" | "tactical") => {
      sendAction("select_secondary_mode", { mode });
    },
    [sendAction],
  );

  const handleFixedSecondariesChange = useCallback(
    (secondaryIds: string[]) => {
      sendAction("select_fixed_secondaries", { secondaryIds });
    },
    [sendAction],
  );

  const handleTogglePainted = useCallback(
    (painted: boolean) => {
      sendAction("set_paint_score", { score: painted ? 10 : 0 });
    },
    [sendAction],
  );

  const handleReady = useCallback(() => {
    sendAction("set_ready", { ready: !myPlayer?.ready });
  }, [sendAction, myPlayer?.ready]);

  const copyInviteCode = () => {
    if (gameState?.inviteCode) {
      void navigator.clipboard.writeText(gameState.inviteCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!gameState) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-3">
          <Spinner size="lg" className="text-primary" />
          <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-primary">
            {connected ? "Loading game" : "Connecting"}
          </p>
        </div>
      </div>
    );
  }

  const hasFaction = !!myPlayer?.factionId;
  const hasDetachment = !!myPlayer?.detachmentId;
  const hasSide = !!myPlayer?.side;
  const hasDisposition = !!myPlayer?.forceDisposition;
  const hasMission = !!myPlayer?.missionId;
  const hasFirstPlayer = (gameState.firstTurnPlayer ?? 0) > 0;
  const hasMode = !!myPlayer?.secondaryMode;
  const fixedSecondaryIds = myPlayer?.fixedSecondaryIds ?? [];
  // Fixed players must pick their set; tactical players are dealt their deck at
  // game start, so the mode choice is enough.
  const hasSecondaries =
    myPlayer?.secondaryMode === "fixed"
      ? fixedSecondaryIds.length === FIXED_SECONDARY_COUNT
      : hasMode;
  // Mission resolution waits on the opponent's disposition, so it is not gated
  // on — readiness only needs the local player's own choices.
  const canReady =
    hasFaction &&
    hasDetachment &&
    hasSide &&
    hasDisposition &&
    hasFirstPlayer &&
    hasMode &&
    hasSecondaries;

  return (
    <div className="relative min-h-screen overflow-hidden bg-background">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            "linear-gradient(var(--primary) 1px, transparent 1px), linear-gradient(90deg, var(--primary) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      />

      <header className="relative z-10 border-b border-border/50 bg-background/60 backdrop-blur-sm">
        <div className="mx-auto flex max-w-2xl items-center justify-between px-4 py-3">
          <div className="flex items-baseline gap-3">
            <span className="font-mono text-base uppercase tracking-[0.3em] text-primary">
              Game Setup
            </span>
          </div>
          <div className="flex items-center gap-2">
            <ThemeSwitcher />
            <ShareSpectateButton gameId={gameState.gameId} />
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={copyInviteCode}
              className="gap-2 font-mono uppercase tracking-widest"
            >
              {copied ? (
                <>
                  <Check className="size-3.5" />
                  Copied!
                </>
              ) : (
                <>
                  <Copy className="size-3.5" />
                  Invite: {gameState.inviteCode}
                </>
              )}
            </Button>
          </div>
        </div>
        {!opponent && (
          <p className="border-t border-border/50 bg-amber-500/5 px-4 py-2 text-center font-mono text-[10px] uppercase tracking-widest text-amber-400">
            Waiting for opponent to join...
          </p>
        )}
        {reconnecting && (
          <div
            role="status"
            aria-live="polite"
            className="flex items-center justify-center gap-2 border-t border-amber-500/40 bg-amber-500/10 px-4 py-2 font-mono text-[10px] uppercase tracking-widest text-amber-300"
          >
            <Spinner className="text-amber-300" />
            Reconnecting to server...
          </div>
        )}
      </header>

      <main className="relative z-0 mx-auto max-w-2xl space-y-5 px-4 py-6">
        <HUDFrame label="Your Faction">
          <FactionPicker
            factions={factions}
            selectedId={myPlayer?.factionId || ""}
            onSelect={handleSelectFaction}
          />
        </HUDFrame>

        {hasFaction && (
          <HUDFrame label="Detachment">
            <DetachmentPicker
              detachments={detachments}
              selectedId={myPlayer?.detachmentId || ""}
              onSelect={handleSelectDetachment}
            />
          </HUDFrame>
        )}

        {hasDetachment && (
          <HUDFrame label="Army Painted">
            <ArmyPaintedToggle
              painted={(myPlayer?.vpPaint ?? 0) > 0}
              onToggle={handleTogglePainted}
            />
          </HUDFrame>
        )}

        {hasDetachment && (
          <HUDFrame label="Your Side">
            <SidePicker selected={myPlayer?.side || ""} onSelect={handleSelectSide} />
          </HUDFrame>
        )}

        {hasSide && (
          <HUDFrame label="Force Disposition">
            <ForceDispositionPicker
              dispositions={availableDispositions}
              selectedId={myPlayer?.forceDisposition || ""}
              onSelect={handleSelectDisposition}
            />
          </HUDFrame>
        )}

        {hasDisposition && (
          <HUDFrame label="Your Primary Mission">
            {hasMission ? (
              <div className="space-y-2 py-1">
                <p className="text-sm font-medium text-foreground">{myPlayer?.missionName}</p>
                <div className="flex flex-wrap gap-2">
                  <Badge
                    variant="outline"
                    className="border-primary/40 bg-primary/10 font-mono uppercase tracking-widest text-primary"
                  >
                    {gameState.vpPerRoundCap} VP / Round
                  </Badge>
                  <Badge
                    variant="outline"
                    className="border-primary/40 bg-primary/10 font-mono uppercase tracking-widest text-primary"
                  >
                    {gameState.vpPerGameCap} VP / Game
                  </Badge>
                </div>
                <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                  Resolved from both players' force dispositions
                </p>
              </div>
            ) : (
              <p className="py-1 font-mono text-[10px] uppercase tracking-widest text-amber-400">
                Waiting for opponent's disposition to resolve your mission...
              </p>
            )}
          </HUDFrame>
        )}

        {hasDisposition && myPlayer && (
          <HUDFrame label="Who Goes First?">
            <FirstPlayerPicker
              myPlayerNumber={myPlayer.playerNumber}
              myUsername={myPlayer.username}
              opponentUsername={opponent?.username}
              selected={gameState.firstTurnPlayer ?? 0}
              onSelect={handleSelectFirstPlayer}
              onRandom={handleRandomFirstPlayer}
            />
          </HUDFrame>
        )}

        {hasFirstPlayer && (
          <HUDFrame label="Secondary Missions">
            <div className="space-y-4">
              <SecondaryModePicker
                mode={myPlayer?.secondaryMode || ""}
                onModeChange={handleModeChange}
              />
              {myPlayer?.secondaryMode === "fixed" && (
                <FixedSecondaryPicker
                  cards={secondaryCards}
                  selectedIds={fixedSecondaryIds}
                  max={FIXED_SECONDARY_COUNT}
                  onChange={handleFixedSecondariesChange}
                />
              )}
            </div>
          </HUDFrame>
        )}

        {opponent && (
          <HUDFrame label={`Opponent: ${opponent.username}`}>
            <div className="space-y-1 py-1">
              <p className="text-sm text-foreground/90">
                {opponent.factionName || "Selecting faction..."}
                {opponent.detachmentName && ` - ${opponent.detachmentName}`}
              </p>
              <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                {opponent.side ? `Side: ${opponent.side}` : "Side: pending"}
                {opponent.forceDispositionName ? ` · ${opponent.forceDispositionName}` : ""}
              </p>
              <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Army: {(opponent.vpPaint ?? 0) > 0 ? "Painted (+10 VP)" : "Not painted"}
              </p>
              <Badge
                variant={opponent.ready ? "default" : "outline"}
                className={cn(
                  "font-mono uppercase tracking-widest",
                  opponent.ready
                    ? "border-emerald-500/50 bg-emerald-500/10 text-emerald-400"
                    : "border-amber-500/50 text-amber-400",
                )}
              >
                {opponent.ready ? "Ready" : "Not ready"}
              </Badge>
            </div>
          </HUDFrame>
        )}

        <Button
          type="button"
          onClick={handleReady}
          disabled={!canReady}
          size="lg"
          className={cn(
            "w-full gap-2 font-mono uppercase tracking-widest",
            myPlayer?.ready &&
              "bg-emerald-600 text-white hover:bg-emerald-700 focus-visible:ring-emerald-500/30",
          )}
        >
          <Swords className="size-4" />
          {myPlayer?.ready ? "Ready! (click to unready)" : "Ready Up"}
        </Button>
      </main>
    </div>
  );
}
