import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { History, LogOut, Plus, Swords, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";
import { ErrorBanner } from "../components/ErrorBanner";
import { useAuth } from "../hooks/useAuth";
import { GameSummary } from "../types/game";
import { ConfirmModal } from "../components/game/ConfirmModal";
import { useGameList } from "../hooks/queries/useGamesQueries";
import { useCreateGame, useJoinGame, useHideGame } from "../hooks/queries/useGameMutations";

export function LobbyPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [joinCode, setJoinCode] = useState("");
  const [gameToRemove, setGameToRemove] = useState<GameSummary | null>(null);

  const { data: games = [] } = useGameList();
  const createGame = useCreateGame();
  const joinGame = useJoinGame();
  const hideGame = useHideGame();

  const loading = createGame.isPending || joinGame.isPending;
  const error = createGame.error
    ? "Failed to create game"
    : joinGame.error
      ? "Invalid invite code"
      : hideGame.error
        ? "Failed to remove game"
        : "";

  const handleCreate = () => {
    createGame.mutate(undefined, {
      onSuccess: ({ id }) => navigate(`/game/${id}/setup`),
    });
  };

  const handleJoin = () => {
    if (!joinCode.trim()) return;
    joinGame.mutate(joinCode.trim().toUpperCase(), {
      onSuccess: ({ id }) => navigate(`/game/${id}/setup`),
    });
  };

  const handleRemoveGame = () => {
    if (!gameToRemove) return;
    hideGame.mutate(gameToRemove.id, {
      onSettled: () => setGameToRemove(null),
    });
  };

  const statusVariant = (status: string) => {
    if (status === "active") return "default" as const;
    if (status === "setup") return "secondary" as const;
    return "outline" as const;
  };

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
            Lobby
          </span>
        </div>
        <div className="flex items-center gap-2">
          {user && (
            <span className="hidden font-mono text-[10px] uppercase tracking-widest text-muted-foreground sm:inline">
              {user.username}
            </span>
          )}
          <ThemeSwitcher />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate("/history")}
            className="gap-1.5 font-mono text-[10px] uppercase tracking-widest"
          >
            <History className="size-3.5" />
            History
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={logout}
            className="gap-1.5 font-mono text-[10px] uppercase tracking-widest"
          >
            <LogOut className="size-3.5" />
            Logout
          </Button>
        </div>
      </header>

      <main className="relative z-0 mx-auto max-w-2xl space-y-6 px-4 py-8">
        {error && <ErrorBanner message={error} />}

        <div className="grid gap-6 sm:grid-cols-2">
          <HUDFrame label="New Game">
            <div className="space-y-4 py-2">
              <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Set up a new game
              </p>
              <Button
                onClick={handleCreate}
                disabled={loading}
                className="w-full gap-2 font-mono uppercase tracking-widest"
              >
                <Plus className="size-4" />
                Create Game
              </Button>
            </div>
          </HUDFrame>

          <HUDFrame label="Join Game">
            <div className="space-y-4 py-2">
              <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Enter 6-character invite
              </p>
              <div className="flex gap-2">
                <Input
                  type="text"
                  placeholder="ABC123"
                  value={joinCode}
                  onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleJoin();
                  }}
                  maxLength={6}
                  className="font-mono uppercase tracking-[0.3em]"
                />
                <Button
                  onClick={handleJoin}
                  disabled={loading || !joinCode.trim()}
                  className="gap-2 font-mono uppercase tracking-widest"
                >
                  <Swords className="size-4" />
                  Join
                </Button>
              </div>
            </div>
          </HUDFrame>
        </div>

        {games.length > 0 && (
          <HUDFrame label="Active Games">
            <div className="space-y-3 py-2">
              {games.map((game, i) => (
                <div key={game.id}>
                  {i > 0 && <Separator className="mb-3" />}
                  <div className="flex items-stretch gap-2">
                    <button
                      onClick={() =>
                        navigate(
                          game.status === "setup" ? `/game/${game.id}/setup` : `/game/${game.id}`,
                        )
                      }
                      className="flex-1 rounded-sm border border-border/50 bg-background/40 p-3 text-left transition-colors hover:border-primary/50 hover:bg-primary/5"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <span className="font-medium text-foreground">
                          {game.missionName || "No mission selected"}
                        </span>
                        <Badge
                          variant={statusVariant(game.status)}
                          className="font-mono uppercase tracking-widest"
                        >
                          {game.status}
                        </Badge>
                      </div>
                      <div className="mt-1 font-mono text-xs text-muted-foreground">
                        {(game.players ?? []).map((p) => p.username).join(" vs ") ||
                          "Awaiting opponent"}
                      </div>
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
                </div>
              ))}
            </div>
          </HUDFrame>
        )}
      </main>

      {gameToRemove && (
        <ConfirmModal
          title="Remove Game"
          message="Are you sure you want to remove this game? It will no longer appear in your game list. This cannot be undone."
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
