import { Mission } from "../../types/mission";

interface Props {
  missions: Mission[];
  selectedId: string;
  onSelect: (mission: Mission) => void;
  onDrawRandom: () => void;
}

export function MissionPicker({ missions, selectedId, onSelect, onDrawRandom }: Props) {
  const selected = missions.find((m) => m.id === selectedId);

  return (
    <div className="space-y-3">
      <div className="flex gap-2">
        <select
          value={selectedId}
          onChange={(e) => {
            const m = missions.find((m) => m.id === e.target.value);
            if (m) onSelect(m);
          }}
          className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white"
        >
          <option value="">Select a mission...</option>
          {missions.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>
        <button
          onClick={onDrawRandom}
          className="bg-indigo-600 hover:bg-indigo-700 px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap"
        >
          Random
        </button>
      </div>
      {selected && (
        <div className="bg-gray-800 rounded-lg p-3 text-sm text-gray-300">
          {selected.lore && <p className="italic text-gray-400 mb-2">{selected.lore}</p>}
          <p className="whitespace-pre-wrap">{selected.description}</p>
        </div>
      )}
    </div>
  );
}
