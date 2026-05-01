import { useState } from "react";
import {
  type NormalizedEvent,
  type PlayerInfoMap,
  formatEvent,
  isHighlightEvent,
} from "./eventFormatting";
import { PlayerAvatar } from "./PlayerAvatar";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Props {
  events: NormalizedEvent[];
  defaultFilter?: "highlights" | "all";
  players?: PlayerInfoMap;
}

export function EventTimeline({ events, defaultFilter = "highlights", players }: Props) {
  const [filter, setFilter] = useState(defaultFilter);

  const filtered = filter === "highlights" ? events.filter(isHighlightEvent) : events;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          Events
        </h3>
        <Select value={filter} onValueChange={(value) => setFilter(value as "highlights" | "all")}>
          <SelectTrigger size="sm" className="font-mono text-[10px] uppercase tracking-widest">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="highlights">Highlights</SelectItem>
            <SelectItem value="all">All Events</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <ScrollArea className="h-80 rounded-sm border border-border/60 bg-background/40 p-3">
        {filtered.length === 0 ? (
          <p className="text-center font-mono text-xs text-muted-foreground">No events.</p>
        ) : (
          <div className="space-y-1">
            {[...filtered].reverse().map((event, i) => {
              const info = event.playerNumber ? players?.[event.playerNumber] : undefined;
              return (
                <div
                  key={i}
                  className="flex items-center gap-2 font-mono text-xs text-muted-foreground"
                >
                  {event.round != null && (
                    <span className="shrink-0 text-primary/60">R{event.round}</span>
                  )}
                  {info && (
                    <PlayerAvatar avatarUrl={info.avatarUrl} username={info.username} size="xs" />
                  )}
                  <span className="text-foreground/80">{formatEvent(event, players)}</span>
                </div>
              );
            })}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}
