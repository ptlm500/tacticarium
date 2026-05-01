import { Undo2 } from "lucide-react";
import { PrimaryScoringSlot, PRIMARY_SCORING_SLOT_LABELS } from "../../types/scoring";
import { Button } from "@/components/ui/button";

interface Props {
  scoredSlots: Record<string, Record<string, Record<string, number>>>;
  onUndo: (round: number, scoringSlot: PrimaryScoringSlot, ruleLabel: string) => void;
}

interface Entry {
  round: number;
  slot: PrimaryScoringSlot;
  ruleLabel: string;
  delta: number;
}

export function PrimaryScoreHistory({ scoredSlots, onUndo }: Props) {
  const entries: Entry[] = [];
  for (const [roundStr, slots] of Object.entries(scoredSlots)) {
    const round = parseInt(roundStr, 10);
    if (Number.isNaN(round)) continue;
    for (const [slot, rules] of Object.entries(slots)) {
      if (!(slot in PRIMARY_SCORING_SLOT_LABELS)) continue;
      for (const [ruleLabel, delta] of Object.entries(rules)) {
        entries.push({ round, slot: slot as PrimaryScoringSlot, ruleLabel, delta });
      }
    }
  }

  if (entries.length === 0) return null;

  entries.sort(
    (a, b) =>
      a.round - b.round || a.slot.localeCompare(b.slot) || a.ruleLabel.localeCompare(b.ruleLabel),
  );

  return (
    <div className="space-y-1">
      <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Primary Scores
      </h3>
      <ul className="space-y-1">
        {entries.map(({ round, slot, ruleLabel, delta }) => (
          <li
            key={`${round}-${slot}-${ruleLabel}`}
            className="flex items-center justify-between rounded-sm border border-border/40 bg-background/40 px-2 py-1 text-xs"
          >
            <span className="text-foreground/80">
              R{round} · {PRIMARY_SCORING_SLOT_LABELS[slot]}
              {ruleLabel ? ` · ${ruleLabel}` : ""} · +{delta}
            </span>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => onUndo(round, slot, ruleLabel)}
              className="h-6 gap-1 px-2 text-xs text-destructive hover:bg-destructive/10 hover:text-destructive"
            >
              <Undo2 className="size-3" />
              Undo
            </Button>
          </li>
        ))}
      </ul>
    </div>
  );
}
