interface Props {
  /** The current viewer's player number (1 or 2). */
  myPlayerNumber: number;
  /** Username shown for "me". */
  myUsername: string;
  /** Username of the opponent, if they have joined. */
  opponentUsername?: string;
  /** Currently selected first-turn player (0 = not yet chosen). */
  selected: number;
  onSelect: (playerNumber: 1 | 2) => void;
  onRandom: () => void;
}

export function FirstPlayerPicker({
  myPlayerNumber,
  myUsername,
  opponentUsername,
  selected,
  onSelect,
  onRandom,
}: Props) {
  const opponentPlayerNumber = (myPlayerNumber === 1 ? 2 : 1) as 1 | 2;
  const selectedMe = selected === myPlayerNumber;
  const selectedOpponent = selected !== 0 && selected === opponentPlayerNumber;

  const baseBtn = "flex-1 px-3 py-2 rounded-lg text-sm font-medium border transition-colors";
  const idle = "bg-gray-800 border-gray-700 hover:bg-gray-700 text-white";
  const active = "bg-indigo-600 border-indigo-500 text-white";

  return (
    <div className="space-y-2">
      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => onSelect(myPlayerNumber as 1 | 2)}
          className={`${baseBtn} ${selectedMe ? active : idle}`}
        >
          {myUsername} (you)
        </button>
        <button
          type="button"
          onClick={() => onSelect(opponentPlayerNumber)}
          disabled={!opponentUsername}
          className={`${baseBtn} ${
            selectedOpponent ? active : idle
          } disabled:opacity-50 disabled:cursor-not-allowed`}
        >
          {opponentUsername ?? "Opponent"}
        </button>
      </div>
      <button
        type="button"
        onClick={onRandom}
        disabled={!opponentUsername}
        className="w-full bg-gray-800 hover:bg-gray-700 border border-gray-700 px-3 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Random
      </button>
      {selected === 0 && (
        <p className="text-xs text-yellow-400">Pick who goes first before readying up.</p>
      )}
    </div>
  );
}
