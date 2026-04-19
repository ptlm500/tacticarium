import { PrimaryScoringSlot, PRIMARY_SCORING_SLOT_LABELS } from "../../types/scoring";

interface Props {
  scoredSlots: Record<string, Record<string, number>>;
  onUndo: (round: number, scoringSlot: PrimaryScoringSlot) => void;
}

export function PrimaryScoreHistory({ scoredSlots, onUndo }: Props) {
  const entries: { round: number; slot: PrimaryScoringSlot; delta: number }[] = [];
  for (const [roundStr, slots] of Object.entries(scoredSlots ?? {})) {
    const round = parseInt(roundStr, 10);
    if (Number.isNaN(round)) continue;
    for (const [slot, delta] of Object.entries(slots)) {
      if (
        slot === "end_of_command_phase" ||
        slot === "end_of_battle_round" ||
        slot === "end_of_turn"
      ) {
        entries.push({ round, slot, delta });
      }
    }
  }

  if (entries.length === 0) return null;

  entries.sort((a, b) => a.round - b.round || a.slot.localeCompare(b.slot));

  return (
    <div className="space-y-1">
      <h3 className="text-xs font-semibold text-gray-400 uppercase">Primary Scores — Undo</h3>
      <ul className="space-y-1">
        {entries.map(({ round, slot, delta }) => (
          <li
            key={`${round}-${slot}`}
            className="flex items-center justify-between bg-gray-800 rounded px-2 py-1 text-xs"
          >
            <span className="text-gray-200">
              R{round} · {PRIMARY_SCORING_SLOT_LABELS[slot]} · +{delta}
            </span>
            <button
              onClick={() => onUndo(round, slot)}
              className="bg-red-900 hover:bg-red-800 text-red-100 px-2 py-0.5 rounded"
            >
              Undo
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
