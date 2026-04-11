interface Props {
  deckSize: number;
  activeCount: number;
  onDraw: () => void;
}

export function TacticalDrawReminder({ deckSize, activeCount, onDraw }: Props) {
  const canDraw = activeCount < 2 && deckSize > 0;
  return (
    <div className="bg-amber-900/40 border border-amber-700 rounded-lg p-3">
      <h3 className="text-sm font-semibold text-amber-200">Draw Tactical Secondaries</h3>
      <p className="text-xs text-amber-300 mt-1">
        {canDraw
          ? `You have ${activeCount}/2 active secondaries. Draw to fill your active slots.`
          : activeCount >= 2
            ? "You already have 2 active secondaries."
            : "Deck is empty."}
      </p>
      {canDraw && (
        <button
          onClick={onDraw}
          className="mt-2 bg-amber-700 hover:bg-amber-600 text-white text-xs px-3 py-2 rounded transition-colors"
        >
          Draw Secondaries ({deckSize} remaining)
        </button>
      )}
    </div>
  );
}
