import { GameEvent } from "../../types/game";
import { formatEvent, normalizeWsEvent } from "./eventFormatting";
import { ScrollArea } from "@/components/ui/scroll-area";

interface Props {
  events: GameEvent[];
}

export function GameLog({ events }: Props) {
  return (
    <ScrollArea className="mt-2 h-60 rounded-sm border border-border/60 bg-background/40 p-3">
      {events.length === 0 ? (
        <p className="text-center text-sm text-muted-foreground">No events yet.</p>
      ) : (
        <div className="space-y-1">
          {[...events].reverse().map((event, i) => (
            <div key={i} className="flex gap-2 font-mono text-xs text-muted-foreground">
              {event.round && <span className="shrink-0 text-primary/60">R{event.round}</span>}
              <span className="text-foreground/80">{formatEvent(normalizeWsEvent(event))}</span>
            </div>
          ))}
        </div>
      )}
    </ScrollArea>
  );
}
