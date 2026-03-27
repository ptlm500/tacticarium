import { GameEvent } from '../../types/game';

interface Props {
  events: GameEvent[];
}

function formatEvent(event: GameEvent): string {
  const player = event.playerNumber ? `P${event.playerNumber}` : '';

  switch (event.eventType) {
    case 'phase_advance':
      return `${player} advanced to ${event.data?.to || 'next'} phase`;
    case 'cp_gain':
      return `${player} gained ${event.data?.amount || 1} CP`;
    case 'cp_adjust':
      return `${player} adjusted CP by ${event.data?.delta}`;
    case 'stratagem_used':
      return `${player} used ${event.data?.stratagemName} (${event.data?.cpSpent} CP)`;
    case 'vp_primary_score':
    case 'vp_secondary_score':
    case 'vp_gambit_score':
      return `${player} scored ${event.data?.delta} ${event.data?.category} VP`;
    case 'game_start':
      return 'Game started!';
    case 'game_end':
      return `Game ended (${event.data?.reason})`;
    case 'player_concede':
      return `${player} conceded`;
    default:
      return `${player} ${event.eventType}`;
  }
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
              {event.round && (
                <span className="text-gray-600 shrink-0">R{event.round}</span>
              )}
              <span>{formatEvent(event)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
