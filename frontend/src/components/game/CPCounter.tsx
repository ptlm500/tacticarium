interface Props {
  cp: number;
  onAdjust: (delta: number) => void;
}

export function CPCounter({ cp, onAdjust }: Props) {
  return (
    <div className="text-center">
      <p className="text-xs text-gray-400 mb-1">Command Points</p>
      <div className="flex items-center justify-center gap-3">
        <button
          onClick={() => onAdjust(-1)}
          disabled={cp <= 0}
          className="w-10 h-10 rounded-full bg-gray-700 hover:bg-gray-600 disabled:opacity-30 text-xl font-bold transition-colors"
        >
          -
        </button>
        <span className="text-3xl font-bold w-12 text-center">{cp}</span>
        <button
          onClick={() => onAdjust(1)}
          className="w-10 h-10 rounded-full bg-gray-700 hover:bg-gray-600 text-xl font-bold transition-colors"
        >
          +
        </button>
      </div>
    </div>
  );
}
