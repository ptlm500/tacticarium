import { Shuffle } from "lucide-react";
import { MissionRule } from "../../types/mission";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Props {
  rules: MissionRule[];
  selectedId: string;
  onSelect: (rule: MissionRule) => void;
  onDrawRandom: () => void;
}

export function TwistPicker({ rules, selectedId, onSelect, onDrawRandom }: Props) {
  const selected = rules.find((r) => r.id === selectedId);

  return (
    <div className="space-y-3">
      <div className="flex gap-2">
        <Select
          value={selectedId}
          onValueChange={(value) => {
            const r = rules.find((r) => r.id === value);
            if (r) onSelect(r);
          }}
        >
          <SelectTrigger className="flex-1">
            <SelectValue placeholder="Select a twist..." />
          </SelectTrigger>
          <SelectContent>
            {rules.map((r) => (
              <SelectItem key={r.id} value={r.id}>
                {r.name}
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
