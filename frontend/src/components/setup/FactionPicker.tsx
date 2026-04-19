import { Faction } from "../../types/faction";
import { cn } from "@/lib/utils";

interface Props {
  factions: Faction[];
  selectedId: string;
  onSelect: (faction: Faction) => void;
}

export function FactionPicker({ factions, selectedId, onSelect }: Props) {
  return (
    <div className="grid max-h-60 grid-cols-2 gap-2 overflow-y-auto pr-1">
      {factions.map((faction) => {
        const active = faction.id === selectedId;
        return (
          <button
            key={faction.id}
            type="button"
            onClick={() => onSelect(faction)}
            className={cn(
              "rounded-sm border p-3 text-left text-sm transition-colors",
              active
                ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
            )}
          >
            {faction.name}
          </button>
        );
      })}
    </div>
  );
}
