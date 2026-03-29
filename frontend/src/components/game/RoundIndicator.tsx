interface Props {
  currentRound: number;
  currentTurn: number;
  maxRounds: number;
}

export function RoundIndicator({ currentRound, currentTurn, maxRounds }: Props) {
  return (
    <div className="flex flex-col items-center gap-1">
      <div className="flex items-center gap-2 justify-center">
        {Array.from({ length: maxRounds }, (_, i) => {
          const round = i + 1;
          const isActive = round === currentRound;
          const isPast = round < currentRound;
          return (
            <div
              key={round}
              className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
                isActive
                  ? 'bg-indigo-600 text-white'
                  : isPast
                  ? 'bg-indigo-900 text-indigo-300'
                  : 'bg-gray-800 text-gray-500'
              }`}
            >
              {round}
            </div>
          );
        })}
      </div>
      <span className="text-xs text-gray-400">
        Round {currentRound} &middot; Turn {currentTurn} of 2
      </span>
    </div>
  );
}
