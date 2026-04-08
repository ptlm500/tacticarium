import { GameEvent } from "../../types/game";
import { formatEvent, normalizeWsEvent } from "./eventFormatting";

interface Props {
  events: GameEvent[];
}

export function GameLog({ events }: Props) {
  return (
    <div className="mt-2 bg-gray-800/50 rounded-lg p-3 max-h-60 overflow-y-auto">
      {events.length === 0 ? (
        <p className="text-gray-500 text-sm text-center">No events yet.</p>
      ) : (
        <div className="space-y-1">
          {[...events].reverse().map((event, i) => (
            <div key={i} className="text-xs text-gray-400 flex gap-2">
              {event.round && <span className="text-gray-600 shrink-0">R{event.round}</span>}
              <span>{formatEvent(normalizeWsEvent(event))}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
