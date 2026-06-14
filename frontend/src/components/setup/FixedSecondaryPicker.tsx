import { MissionCard } from "../../types/mission";
import { cn } from "@/lib/utils";

interface Props {
  cards: MissionCard[];
  selectedIds: string[];
  /** Maximum number of cards a fixed-mode player may keep for the game. */
  max: number;
  onChange: (ids: string[]) => void;
}

export function FixedSecondaryPicker({ cards, selectedIds, max, onChange }: Props) {
  const toggle = (id: string) => {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((s) => s !== id));
    } else if (selectedIds.length < max) {
      onChange([...selectedIds, id]);
    }
  };

  return (
    <div className="space-y-2">
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Choose {max} fixed secondaries ({selectedIds.length}/{max})
      </p>
      <div className="max-h-72 space-y-2 overflow-y-auto pr-1">
        {cards.map((c) => {
          const active = selectedIds.includes(c.id);
          const atCapacity = !active && selectedIds.length >= max;
          return (
            <button
              key={c.id}
              type="button"
              onClick={() => toggle(c.id)}
              disabled={atCapacity}
              className={cn(
                "w-full rounded-sm border p-3 text-left transition-colors",
                active
                  ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                  : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
                atCapacity &&
                  "cursor-not-allowed opacity-40 hover:border-border/60 hover:bg-background/40",
              )}
            >
              <span className="block text-sm font-medium">{c.name}</span>
              {c.text && (
                <span className="mt-1 block text-xs leading-snug text-muted-foreground">
                  {c.text}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
