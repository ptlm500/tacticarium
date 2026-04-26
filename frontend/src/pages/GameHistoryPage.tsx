import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, X } from "lucide-react";
import { GameSummary } from "../types/game";
import { useAuth } from "../hooks/useAuth";
import { ConfirmModal } from "../components/game/ConfirmModal";
import { useGameHistory, useUserStats } from "../hooks/queries/useHistoryQueries";
import { useFactions } from "../hooks/queries/useFactionQueries";
import { useHideGame } from "../hooks/queries/useGameMutations";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";
import { ErrorBanner } from "../components/ErrorBanner";

const ALL_FACTIONS = "__all__";

export function GameHistoryPage() {
  const navigate = useNavigate();
  const { user } = useAuth();

  const [myFaction, setMyFaction] = useState("");
  const [opponentFaction, setOpponentFaction] = useState("");
  const [gameToRemove, setGameToRemove] = useState<GameSummary | null>(null);

  const filters =
    myFaction || opponentFaction
      ? { myFaction: myFaction || undefined, opponentFaction: opponentFaction || undefined }
      : undefined;

  const { data: games = [], isPending: loading } = useGameHistory(filters);
  const { data: stats } = useUserStats();
  const { data: factions = [] } = useFactions();
  const hideGame = useHideGame();

  const error = hideGame.error?.message || "";

  const handleRemoveGame = () => {
    if (!gameToRemove) return;
    hideGame.mutate(gameToRemove.id, {
      onSettled: () => setGameToRemove(null),
    });
  };

  const resultBadge = (game: GameSummary) => {
    const isWinner = game.winnerId === user?.id;
    const isDraw = !game.winnerId && game.status === "completed";
    const isAbandoned = game.status === "abandoned";
    if (isAbandoned) {
      return (
        <Badge
          variant="outline"
          className="border-amber-500/40 bg-amber-500/10 font-mono uppercase tracking-widest text-amber-300"
        >
          Abandoned
        </Badge>
      );
    }
    if (isDraw) {
      return (
        <Badge variant="outline" className="font-mono uppercase tracking-widest">
          Draw
        </Badge>
      );
    }
    if (isWinner) {
      return (
        <Badge
          variant="outline"
          className="border-emerald-500/40 bg-emerald-500/10 font-mono uppercase tracking-widest text-emerald-300"
        >
          Won
        </Badge>
      );
    }
    return (
      <Badge
        variant="outline"
        className="border-destructive/40 bg-destructive/10 font-mono uppercase tracking-widest text-destructive"
      >
        Lost
      </Badge>
    );
  };

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
            Battle Archive
          </span>
        </div>
        <div className="flex items-center gap-2">
          <ThemeSwitcher />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate("/")}
            className="gap-1.5 font-mono text-[10px] uppercase tracking-widest"
          >
            <ArrowLeft className="size-3.5" />
            Back
          </Button>
        </div>
      </header>

      <main className="relative z-0 mx-auto max-w-2xl space-y-6 px-4 py-8">
        {error && <ErrorBanner message={error} />}

        {stats && (
          <HUDFrame label="Campaign Stats">
            <div className="space-y-3 py-1">
              <div className="flex flex-wrap justify-center gap-2">
                <Badge
                  variant="outline"
                  className="border-emerald-500/40 bg-emerald-500/10 font-mono uppercase tracking-widest text-emerald-300"
                >
                  {stats.wins}W
                </Badge>
                <Badge
                  variant="outline"
                  className="border-destructive/40 bg-destructive/10 font-mono uppercase tracking-widest text-destructive"
                >
                  {stats.losses}L
                </Badge>
                <Badge variant="outline" className="font-mono uppercase tracking-widest">
                  {stats.draws}D
                </Badge>
                {stats.abandoned > 0 && (
                  <Badge
                    variant="outline"
                    className="border-amber-500/40 bg-amber-500/10 font-mono uppercase tracking-widest text-amber-300"
                  >
                    {stats.abandoned} Abandoned
                  </Badge>
                )}
              </div>
              <div className="text-center font-mono text-xs uppercase tracking-widest text-muted-foreground">
                Avg VP:{" "}
                <span className="text-foreground tabular-nums">{stats.averageVp.toFixed(1)}</span>
              </div>
              {(stats.factionStats ?? []).length > 0 && (
                <>
                  <Separator />
                  <div className="space-y-1">
                    <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                      Most played
                    </p>
                    {(stats.factionStats ?? []).slice(0, 3).map((fs) => (
                      <div key={fs.factionName} className="flex justify-between font-mono text-xs">
                        <span className="text-foreground">{fs.factionName}</span>
                        <span className="text-muted-foreground">
                          {fs.gamesPlayed} game{fs.gamesPlayed !== 1 ? "s" : ""}, {fs.wins} win
                          {fs.wins !== 1 ? "s" : ""}
                        </span>
                      </div>
                    ))}
                  </div>
                </>
              )}
            </div>
          </HUDFrame>
        )}

        {/* Filters */}
        <div className="flex gap-3">
          <Select
            value={myFaction || ALL_FACTIONS}
            onValueChange={(value) => setMyFaction(value === ALL_FACTIONS ? "" : value)}
          >
            <SelectTrigger className="flex-1 font-mono text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_FACTIONS}>My Faction (all)</SelectItem>
              {factions.map((f) => (
                <SelectItem key={f.id} value={f.name}>
                  {f.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            value={opponentFaction || ALL_FACTIONS}
            onValueChange={(value) => setOpponentFaction(value === ALL_FACTIONS ? "" : value)}
          >
            <SelectTrigger className="flex-1 font-mono text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_FACTIONS}>Opponent (all)</SelectItem>
              {factions.map((f) => (
                <SelectItem key={f.id} value={f.name}>
                  {f.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Game List */}
        {loading ? (
          <div className="flex justify-center py-8">
            <div className="flex items-center gap-2">
              <Spinner />
              <span className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
                Loading
              </span>
            </div>
          </div>
        ) : games.length === 0 ? (
          <p className="text-center font-mono text-xs uppercase tracking-widest text-muted-foreground">
            No completed games yet.
          </p>
        ) : (
          <div className="space-y-3">
            {games.map((game) => (
              <div key={game.id} className="flex items-stretch gap-2">
                <button
                  onClick={() => navigate(`/history/${game.id}`)}
                  className="flex-1 rounded-sm border border-border/50 bg-background/40 p-4 text-left transition-colors hover:border-primary/50 hover:bg-primary/5"
                >
                  <div className="mb-2 flex items-center justify-between gap-2">
                    <span className="font-medium text-foreground">
                      {game.missionName || "Unknown Mission"}
                    </span>
                    {resultBadge(game)}
                  </div>
                  <div className="space-y-1 font-mono text-xs">
                    {(game.players ?? []).map((p) => (
                      <div
                        key={p.userId}
                        className={`flex justify-between ${
                          p.userId === user?.id ? "text-foreground" : "text-muted-foreground"
                        }`}
                      >
                        <span>
                          {p.username}
                          {p.factionName && ` (${p.factionName})`}
                        </span>
                        <span className="tabular-nums">{p.totalVp} VP</span>
                      </div>
                    ))}
                  </div>
                  {game.completedAt && (
                    <p className="mt-2 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                      {new Date(game.completedAt).toLocaleDateString()}
                    </p>
                  )}
                </button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setGameToRemove(game)}
                  aria-label="Remove game"
                  title="Remove game"
                  className="self-center text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                >
                  <X className="size-4" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </main>

      {gameToRemove && (
        <ConfirmModal
          title="Remove Game"
          message="Are you sure you want to remove this game from your history? This cannot be undone."
          confirmLabel="Remove"
          cancelLabel="Cancel"
          variant="danger"
          onConfirm={handleRemoveGame}
          onCancel={() => setGameToRemove(null)}
        />
      )}
    </div>
  );
}
