import { Faction } from '../../types/faction';

interface Props {
  factions: Faction[];
  selectedId: string;
  onSelect: (faction: Faction) => void;
}

export function FactionPicker({ factions, selectedId, onSelect }: Props) {
  return (
    <div className="grid grid-cols-2 gap-2 max-h-60 overflow-y-auto">
      {factions.map((faction) => (
        <button
          key={faction.id}
          onClick={() => onSelect(faction)}
          className={`p-3 rounded-lg text-left text-sm transition-colors ${
            faction.id === selectedId
              ? 'bg-indigo-600 text-white border-2 border-indigo-400'
              : 'bg-gray-800 hover:bg-gray-750 border border-gray-700 text-gray-300'
          }`}
        >
          {faction.name}
        </button>
      ))}
    </div>
  );
}
