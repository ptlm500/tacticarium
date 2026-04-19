import { useState } from "react";

interface Props {
  vpPrimary: number;
  vpSecondary: number;
  vpGambit: number;
  vpPaint: number;
  onAdjust: (category: string, delta: number) => void;
}

export function VPCounter({ vpPrimary, vpSecondary, vpGambit, vpPaint, onAdjust }: Props) {
  const [expanded, setExpanded] = useState(false);
  const total = vpPrimary + vpSecondary + vpGambit + vpPaint;

  return (
    <div className="text-center">
      <p className="text-xs text-gray-400 mb-1">Victory Points</p>
      <button onClick={() => setExpanded(!expanded)} className="text-3xl font-bold">
        {total}
      </button>

      {expanded && (
        <div className="mt-3 space-y-2 text-sm">
          <p className="text-xs text-gray-500 italic">Manual adjust (bypasses mission rules)</p>
          <VPRow
            label="Primary"
            value={vpPrimary}
            max={50}
            category="primary"
            onAdjust={onAdjust}
          />
          <VPRow
            label="Secondary"
            value={vpSecondary}
            max={40}
            category="secondary"
            onAdjust={onAdjust}
          />
          <VPRow label="Gambit" value={vpGambit} max={12} category="gambit" onAdjust={onAdjust} />
          <div className="flex items-center justify-between text-gray-400">
            <span>Paint</span>
            <span>{vpPaint}/10</span>
          </div>
        </div>
      )}
    </div>
  );
}

function VPRow({
  label,
  value,
  max,
  category,
  onAdjust,
}: {
  label: string;
  value: number;
  max: number;
  category: string;
  onAdjust: (category: string, delta: number) => void;
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-gray-400">{label}</span>
      <div className="flex items-center gap-2">
        <button
          onClick={() => onAdjust(category, -1)}
          disabled={value <= 0}
          className="w-6 h-6 rounded bg-gray-700 hover:bg-gray-600 disabled:opacity-30 text-xs"
        >
          -
        </button>
        <span className="w-10 text-center">
          {value}/{max}
        </span>
        <button
          onClick={() => onAdjust(category, 1)}
          disabled={value >= max}
          className="w-6 h-6 rounded bg-gray-700 hover:bg-gray-600 disabled:opacity-30 text-xs"
        >
          +
        </button>
      </div>
    </div>
  );
}
