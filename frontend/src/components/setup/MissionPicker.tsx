import { Shuffle } from "lucide-react";
import { Mission } from "../../types/mission";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

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
          className={cn(
            "flex h-9 flex-1 rounded-md border border-input bg-transparent px-3 py-1 text-sm",
            "shadow-xs transition-[color,box-shadow] outline-none",
            "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
            "dark:bg-input/30",
          )}
        >
          <option value="">Select a mission...</option>
          {missions.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>
        <Button
          type="button"
          variant="outline"
          onClick={onDrawRandom}
          className="gap-2 font-mono uppercase tracking-widest"
        >
          <Shuffle className="size-4" />
          Random
        </Button>
      </div>
      {selected && (
        <div className="rounded-sm border border-border/60 bg-background/40 p-3 text-sm text-foreground/90">
          {selected.lore && <p className="mb-2 italic text-muted-foreground">{selected.lore}</p>}
          <p className="whitespace-pre-wrap">{selected.description}</p>
        </div>
      )}
    </div>
  );
}
