import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { PRIMARY_SCORING_SLOT_LABELS, type PrimaryScoringSlot } from "../../types/scoring";
import type { NormalizedEvent } from "./eventFormatting";
import type { VPCategory } from "./vpUtils";

export interface CellSelection {
  playerNumber: number;
  username: string;
  round: number;
  category: VPCategory;
}

interface Props {
  selection: CellSelection | null;
  onClose: () => void;
  events: NormalizedEvent[];
}

interface DetailEntry {
  key: string;
  label: string;
  delta: number;
  reverted?: boolean;
}

export function ScoringDetailModal({ selection, onClose, events }: Props) {
  const entries = selection ? buildEntries(events, selection) : [];
  const total = entries.reduce((sum, e) => sum + (e.reverted ? 0 : e.delta), 0);
  const categoryLabel = selection?.category === "secondary" ? "Secondary" : "Primary";

  return (
    <Dialog open={selection !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest">
            R{selection?.round ?? 0} · {categoryLabel} VP
          </DialogTitle>
          <DialogDescription className="font-mono text-[11px] uppercase tracking-widest">
            {selection?.username ?? ""} — {total} VP this round
          </DialogDescription>
        </DialogHeader>

        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No scoring events recorded.</p>
        ) : (
          <ul className="space-y-1.5">
            {entries.map((entry) => (
              <li
                key={entry.key}
                className="flex items-center justify-between rounded-sm border border-border/40 bg-background/40 px-2 py-1.5 text-xs"
              >
                <span
                  className={
                    entry.reverted ? "text-muted-foreground line-through" : "text-foreground/85"
                  }
                >
                  {entry.label}
                </span>
                <span className={cn("font-mono tabular-nums", deltaClass(entry))}>
                  {entry.delta >= 0 ? `+${entry.delta}` : entry.delta}
                </span>
              </li>
            ))}
          </ul>
        )}
      </DialogContent>
    </Dialog>
  );
}

function deltaClass(entry: DetailEntry): string {
  if (entry.reverted) return "text-muted-foreground line-through";
  if (entry.delta < 0) return "text-destructive";
  return "text-primary";
}

function buildEntries(events: NormalizedEvent[], selection: CellSelection): DetailEntry[] {
  const { playerNumber, round, category } = selection;
  const out: DetailEntry[] = [];
  const revertedKeys = new Set<string>();

  events.forEach((e) => {
    if (e.playerNumber !== playerNumber) return;

    if (
      e.eventType === "vp_primary_score_reverted" &&
      category === "primary" &&
      ((e.data?.revertedRound as number | undefined) ?? e.round) === round
    ) {
      const slot = (e.data?.scoringSlot as string | undefined) ?? "";
      const ruleLabel = (e.data?.scoringRuleLabel as string | undefined) ?? "";
      revertedKeys.add(`${slot}::${ruleLabel}`);
    }
  });

  events.forEach((e, idx) => {
    if (e.playerNumber !== playerNumber) return;
    if ((e.round ?? 0) !== round) return;

    if (e.eventType === "vp_primary_score" && category === "primary") {
      const slot = (e.data?.scoringSlot as string | undefined) ?? "";
      const ruleLabel = (e.data?.scoringRuleLabel as string | undefined) ?? "";
      const delta =
        (e.data?.appliedDelta as number | undefined) ?? (e.data?.delta as number | undefined) ?? 0;
      const slotLabel = isPrimarySlot(slot) ? PRIMARY_SCORING_SLOT_LABELS[slot] : slot || "Score";
      const label = ruleLabel ? `${slotLabel} · ${ruleLabel}` : slotLabel;
      out.push({
        key: `primary-${idx}`,
        label,
        delta,
        reverted: revertedKeys.has(`${slot}::${ruleLabel}`),
      });
    }

    if (e.eventType === "secondary_achieved" && category === "secondary") {
      const name = (e.data?.secondaryName as string | undefined) ?? "Secondary";
      const delta = (e.data?.vpScored as number | undefined) ?? 0;
      out.push({ key: `sec-ach-${idx}`, label: `${name} achieved`, delta });
    }

    if (e.eventType === "vp_secondary_score" && category === "secondary") {
      const delta =
        (e.data?.appliedDelta as number | undefined) ?? (e.data?.delta as number | undefined) ?? 0;
      const name = (e.data?.secondaryName as string | undefined) ?? null;
      out.push({
        key: `sec-${idx}`,
        label: name ? `${name} · score` : "Secondary score",
        delta,
      });
    }

    if (e.eventType === "vp_manual_adjust") {
      const cat = e.data?.category as string | undefined;
      if (cat !== category) return;
      const delta =
        (e.data?.appliedDelta as number | undefined) ?? (e.data?.delta as number | undefined) ?? 0;
      out.push({ key: `manual-${idx}`, label: "Manual adjust", delta });
    }
  });

  return out;
}

function isPrimarySlot(slot: string): slot is PrimaryScoringSlot {
  return slot in PRIMARY_SCORING_SLOT_LABELS;
}
