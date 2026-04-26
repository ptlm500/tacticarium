import { Shuffle } from "lucide-react";
import { Mission } from "../../types/mission";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

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
        <Select
          value={selectedId}
          onValueChange={(value) => {
            const m = missions.find((m) => m.id === value);
            if (m) onSelect(m);
          }}
        >
          <SelectTrigger className="flex-1">
            <SelectValue placeholder="Select a mission..." />
          </SelectTrigger>
          <SelectContent>
            {missions.map((m) => (
              <SelectItem key={m.id} value={m.id}>
                {m.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
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
