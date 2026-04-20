import { useState } from "react";
import { type NormalizedEvent, formatEvent, isHighlightEvent } from "./eventFormatting";
import { ScrollArea } from "@/components/ui/scroll-area";

interface Props {
  events: NormalizedEvent[];
  defaultFilter?: "highlights" | "all";
}

export function EventTimeline({ events, defaultFilter = "highlights" }: Props) {
  const [filter, setFilter] = useState(defaultFilter);

  const filtered = filter === "highlights" ? events.filter(isHighlightEvent) : events;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          Events
        </h3>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value as "highlights" | "all")}
          className="rounded-md border border-input bg-transparent px-2 py-1 font-mono text-[10px] uppercase tracking-widest text-foreground focus:border-ring focus:outline-none focus:ring-[3px] focus:ring-ring/50"
        >
          <option value="highlights">Highlights</option>
          <option value="all">All Events</option>
        </select>
      </div>
      <ScrollArea className="h-80 rounded-sm border border-border/60 bg-background/40 p-3">
        {filtered.length === 0 ? (
          <p className="text-center font-mono text-xs text-muted-foreground">No events.</p>
        ) : (
          <div className="space-y-1">
            {[...filtered].reverse().map((event, i) => (
              <div key={i} className="flex gap-2 font-mono text-xs text-muted-foreground">
                {event.round != null && (
                  <span className="shrink-0 text-primary/60">R{event.round}</span>
                )}
                <span className="text-foreground/80">{formatEvent(event)}</span>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}
