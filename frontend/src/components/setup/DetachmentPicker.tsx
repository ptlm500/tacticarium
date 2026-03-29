import { Detachment } from "../../types/faction";

interface Props {
  detachments: Detachment[];
  selectedId: string;
  onSelect: (detachment: Detachment) => void;
}

export function DetachmentPicker({ detachments, selectedId, onSelect }: Props) {
  if (detachments.length === 0) {
    return <p className="text-gray-500 text-sm">Loading detachments...</p>;
  }

  return (
    <div className="space-y-2 max-h-48 overflow-y-auto">
      {detachments.map((d) => (
        <button
          key={d.id}
          onClick={() => onSelect(d)}
          className={`w-full p-3 rounded-lg text-left text-sm transition-colors ${
            d.id === selectedId
              ? "bg-indigo-600 text-white border-2 border-indigo-400"
              : "bg-gray-800 hover:bg-gray-750 border border-gray-700 text-gray-300"
          }`}
        >
          {d.name}
        </button>
      ))}
    </div>
  );
}
