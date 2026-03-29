import { Stratagem } from "../../types/faction";

interface Props {
  stratagems: Stratagem[];
  currentCP: number;
  onUse: (stratagem: Stratagem) => void;
}

export function StratagemPanel({ stratagems, currentCP, onUse }: Props) {
  if (stratagems.length === 0) {
    return (
      <div className="mt-2 bg-gray-800/50 rounded-lg p-4 text-gray-500 text-sm text-center">
        No stratagems available for this phase.
      </div>
    );
  }

  return (
    <div className="mt-2 space-y-2 max-h-80 overflow-y-auto">
      {stratagems.map((s) => {
        const canAfford = currentCP >= s.cpCost;
        return (
          <div key={s.id} className="bg-gray-800/50 border border-gray-700 rounded-lg p-3">
            <div className="flex justify-between items-start mb-1">
              <div>
                <h3 className="font-semibold text-sm">{s.name}</h3>
                <p className="text-xs text-gray-400">{s.type}</p>
              </div>
              <span
                className={`text-xs font-bold px-2 py-1 rounded ${
                  s.cpCost === 0 ? "bg-green-900 text-green-300" : "bg-indigo-900 text-indigo-300"
                }`}
              >
                {s.cpCost} CP
              </span>
            </div>
            {s.legend && <p className="text-xs text-gray-400 mb-2">{s.legend}</p>}
            <div className="flex justify-between items-center mt-2">
              <span className="text-xs text-gray-500">
                {s.turn} | {s.phase}
              </span>
              <button
                onClick={() => {
                  if (window.confirm(`Use ${s.name} for ${s.cpCost} CP?`)) {
                    onUse(s);
                  }
                }}
                disabled={!canAfford}
                className="bg-indigo-600 hover:bg-indigo-700 disabled:opacity-30 text-white text-xs font-semibold px-3 py-1 rounded transition-colors"
              >
                Use
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
