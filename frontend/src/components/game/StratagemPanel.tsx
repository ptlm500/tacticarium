import { useState } from "react";
import { Stratagem } from "../../types/faction";

interface Props {
  stratagems: Stratagem[];
  currentCP: number;
  onUse: (stratagem: Stratagem, cpSpent: number) => void;
}

export function StratagemPanel({ stratagems, currentCP, onUse }: Props) {
  const [pending, setPending] = useState<Stratagem | null>(null);
  const [cpInput, setCpInput] = useState<number>(0);

  if (stratagems.length === 0) {
    return (
      <div className="mt-2 bg-gray-800/50 rounded-lg p-4 text-gray-500 text-sm text-center">
        No stratagems available for this phase.
      </div>
    );
  }

  const openPrompt = (s: Stratagem) => {
    setPending(s);
    setCpInput(Math.min(s.cpCost, currentCP));
  };

  const closePrompt = () => setPending(null);

  const confirmUse = () => {
    if (!pending) return;
    const clamped = Math.max(0, Math.min(cpInput, currentCP));
    onUse(pending, clamped);
    setPending(null);
  };

  return (
    <>
      <div className="mt-2 space-y-2 max-h-80 overflow-y-auto">
        {stratagems.map((s) => (
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
                onClick={() => openPrompt(s)}
                className="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold px-3 py-1 rounded transition-colors"
              >
                Use
              </button>
            </div>
          </div>
        ))}
      </div>

      {pending && (
        <div
          className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-label="Confirm stratagem"
        >
          <div className="bg-gray-900 border border-gray-700 rounded-lg p-5 max-w-sm w-full">
            <h3 className="font-semibold text-base mb-1">{pending.name}</h3>
            <p className="text-xs text-gray-400 mb-4">
              Default cost: {pending.cpCost} CP. Adjust the amount to spend if a rule makes this
              stratagem more expensive, cheaper, or free.
            </p>
            <label className="block text-xs text-gray-400 mb-1" htmlFor="strat-cp-input">
              CP to spend (you have {currentCP})
            </label>
            <input
              id="strat-cp-input"
              type="number"
              min={0}
              max={currentCP}
              value={cpInput}
              onChange={(e) => setCpInput(Math.max(0, Number(e.target.value)))}
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm mb-4"
              autoFocus
            />
            {cpInput > currentCP && (
              <p className="text-xs text-red-400 mb-2">
                You only have {currentCP} CP — cannot spend more than that.
              </p>
            )}
            <div className="flex justify-end gap-2">
              <button
                onClick={closePrompt}
                className="px-3 py-1 text-sm rounded bg-gray-700 hover:bg-gray-600 text-white"
              >
                Cancel
              </button>
              <button
                onClick={confirmUse}
                disabled={cpInput > currentCP || cpInput < 0}
                className="px-3 py-1 text-sm rounded bg-indigo-600 hover:bg-indigo-700 disabled:opacity-30 text-white font-semibold"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
