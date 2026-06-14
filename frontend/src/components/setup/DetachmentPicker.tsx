import { Detachment } from "../../types/faction";
import { SelectedDetachment } from "../../types/game";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";

interface Props {
  detachments: Detachment[];
  selectedIds: string[];
  /** Detachment-points budget a player may spend across detachments. */
  maxPoints: number;
  onChange: (selected: SelectedDetachment[]) => void;
}

function pointsOf(d: Detachment): number {
  return d.detachmentPoints ?? 0;
}

export function DetachmentPicker({ detachments, selectedIds, maxPoints, onChange }: Props) {
  if (detachments.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Spinner size="sm" />
        <span className="font-mono text-[10px] uppercase tracking-widest">Loading detachments</span>
      </div>
    );
  }

  const usedPoints = detachments
    .filter((d) => selectedIds.includes(d.id))
    .reduce((sum, d) => sum + pointsOf(d), 0);

  const emit = (ids: string[]) => {
    const selected: SelectedDetachment[] = detachments
      .filter((d) => ids.includes(d.id))
      .map((d) => ({ id: d.id, name: d.name, points: pointsOf(d) }));
    onChange(selected);
  };

  const toggle = (d: Detachment) => {
    if (selectedIds.includes(d.id)) {
      emit(selectedIds.filter((id) => id !== d.id));
    } else if (usedPoints + pointsOf(d) <= maxPoints) {
      emit([...selectedIds, d.id]);
    }
  };

  return (
    <div className="space-y-2">
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Detachment points: {usedPoints}/{maxPoints}
      </p>
      <div className="max-h-60 space-y-2 overflow-y-auto pr-1">
        {detachments.map((d) => {
          const active = selectedIds.includes(d.id);
          const wouldExceed = !active && usedPoints + pointsOf(d) > maxPoints;
          return (
            <button
              key={d.id}
              type="button"
              onClick={() => toggle(d)}
              disabled={wouldExceed}
              className={cn(
                "flex w-full items-center justify-between gap-2 rounded-sm border p-3 text-left text-sm transition-colors",
                active
                  ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                  : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
                wouldExceed &&
                  "cursor-not-allowed opacity-40 hover:border-border/60 hover:bg-background/40",
              )}
            >
              <span>{d.name}</span>
              {pointsOf(d) > 0 ? (
                <span className="shrink-0 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                  {pointsOf(d)} pt{pointsOf(d) === 1 ? "" : "s"}
                </span>
              ) : null}
            </button>
          );
        })}
      </div>
    </div>
  );
}
