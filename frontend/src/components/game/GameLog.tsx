import { useMemo } from "react";
import { GameEvent, GameState } from "../../types/game";
import {
  buildPlayerInfo,
  formatEvent,
  isHighlightEvent,
  normalizeWsEvent,
} from "./eventFormatting";
import { PlayerAvatar } from "./PlayerAvatar";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

interface Props {
  events: GameEvent[];
  players?: GameState["players"];
}

export function GameLog({ events, players }: Props) {
  const items = [...events].reverse();
  const playerInfo = useMemo(() => buildPlayerInfo(players), [players]);

  return (
    <div className="relative mt-2 overflow-hidden rounded border border-primary/30 bg-card/50 backdrop-blur-sm">
      <div className="pointer-events-none absolute left-0 top-0 z-10 h-3 w-3 border-l-2 border-t-2 border-primary/50" />
      <div className="pointer-events-none absolute right-0 top-0 z-10 h-3 w-3 border-r-2 border-t-2 border-primary/50" />
      <div className="pointer-events-none absolute bottom-0 left-0 z-10 h-3 w-3 border-b-2 border-l-2 border-primary/50" />
      <div className="pointer-events-none absolute bottom-0 right-0 z-10 h-3 w-3 border-b-2 border-r-2 border-primary/50" />
      <div className="pointer-events-none absolute inset-0 bg-[repeating-linear-gradient(0deg,transparent,transparent_2px,rgba(0,0,0,0.03)_2px,rgba(0,0,0,0.03)_4px)]" />

      <ScrollArea className="relative h-60">
        <div className="px-4 py-3">
          {items.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No events yet.</p>
          ) : (
            <ol className="relative space-y-2">
              <div className="pointer-events-none absolute bottom-1 left-[5px] top-1 w-px bg-primary/20" />
              {items.map((event, i) => {
                const norm = normalizeWsEvent(event);
                const highlight = isHighlightEvent(norm);
                const info = norm.playerNumber ? playerInfo[norm.playerNumber] : undefined;
                return (
                  <li key={i} className="relative flex items-start gap-3 pl-5">
                    <span
                      aria-hidden
                      className={cn(
                        "absolute left-0 top-[5px] h-[11px] w-[11px] rounded-full border-2",
                        highlight
                          ? "border-primary bg-primary shadow-[0_0_6px_var(--primary)]"
                          : "border-foreground/30 bg-background",
                      )}
                    />
                    <div className="flex flex-1 flex-wrap items-center gap-x-2 font-mono text-xs leading-relaxed">
                      {event.round != null && (
                        <span className="shrink-0 text-primary/60">R{event.round}</span>
                      )}
                      {info && (
                        <PlayerAvatar
                          avatarUrl={info.avatarUrl}
                          username={info.username}
                          size="xs"
                        />
                      )}
                      <span className={cn(highlight ? "text-foreground" : "text-foreground/70")}>
                        {formatEvent(norm, playerInfo)}
                      </span>
                    </div>
                  </li>
                );
              })}
            </ol>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
