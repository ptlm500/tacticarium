import { useState } from "react";
import { type NormalizedEvent, formatEvent, isHighlightEvent } from "./eventFormatting";

interface Props {
  events: NormalizedEvent[];
  defaultFilter?: "highlights" | "all";
}

export function EventTimeline({ events, defaultFilter = "highlights" }: Props) {
  const [filter, setFilter] = useState(defaultFilter);

  const filtered = filter === "highlights" ? events.filter(isHighlightEvent) : events;

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-semibold text-gray-400">Event Timeline</h3>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value as "highlights" | "all")}
          className="bg-gray-700 text-gray-300 text-xs px-2 py-1 rounded border border-gray-600"
        >
          <option value="highlights">Highlights</option>
          <option value="all">All Events</option>
        </select>
      </div>
      <div className="bg-gray-800/50 rounded-lg p-3 max-h-80 overflow-y-auto">
        {filtered.length === 0 ? (
          <p className="text-gray-500 text-sm text-center">No events.</p>
        ) : (
          <div className="space-y-1">
            {[...filtered].reverse().map((event, i) => (
              <div key={i} className="text-xs text-gray-400 flex gap-2">
                {event.round != null && (
                  <span className="text-gray-600 shrink-0">R{event.round}</span>
                )}
                <span>{formatEvent(event)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
