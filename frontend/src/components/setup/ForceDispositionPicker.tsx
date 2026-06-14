import { ForceDisposition } from "../../types/mission";
import { cn } from "@/lib/utils";

interface Props {
  dispositions: ForceDisposition[];
  selectedId: string;
  onSelect: (disposition: ForceDisposition) => void;
}

export function ForceDispositionPicker({ dispositions, selectedId, onSelect }: Props) {
  if (dispositions.length === 0) {
    return (
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        No force dispositions available
      </p>
    );
  }

  return (
    <div className="max-h-72 space-y-2 overflow-y-auto pr-1">
      {dispositions.map((d) => {
        const active = d.id === selectedId;
        return (
          <button
            key={d.id}
            type="button"
            onClick={() => onSelect(d)}
            className={cn(
              "w-full rounded-sm border p-3 text-left transition-colors",
              active
                ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
            )}
          >
            <span className="block text-sm font-medium">{d.name}</span>
            {d.text && (
              <span className="mt-1 block text-xs leading-snug text-muted-foreground">
                {d.text}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}
