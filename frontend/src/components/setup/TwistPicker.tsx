import { MissionRule } from '../../types/mission';

interface Props {
  rules: MissionRule[];
  selectedId: string;
  onSelect: (rule: MissionRule) => void;
  onDrawRandom: () => void;
}

export function TwistPicker({ rules, selectedId, onSelect, onDrawRandom }: Props) {
  const selected = rules.find((r) => r.id === selectedId);

  return (
    <div className="space-y-3">
      <div className="flex gap-2">
        <select
          value={selectedId}
          onChange={(e) => {
            const r = rules.find((r) => r.id === e.target.value);
            if (r) onSelect(r);
          }}
          className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white"
        >
          <option value="">Select a twist...</option>
          {rules.map((r) => (
            <option key={r.id} value={r.id}>
              {r.name}
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
          {selected.lore && (
            <p className="italic text-gray-400 mb-2">{selected.lore}</p>
          )}
          <p className="whitespace-pre-wrap">{selected.description}</p>
        </div>
      )}
    </div>
  );
}
