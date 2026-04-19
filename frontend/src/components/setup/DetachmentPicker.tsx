import { Detachment } from "../../types/faction";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";

interface Props {
  detachments: Detachment[];
  selectedId: string;
  onSelect: (detachment: Detachment) => void;
}

export function DetachmentPicker({ detachments, selectedId, onSelect }: Props) {
  if (detachments.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Spinner size="sm" />
        <span className="font-mono text-[10px] uppercase tracking-widest">Loading detachments</span>
      </div>
    );
  }

  return (
    <div className="max-h-48 space-y-2 overflow-y-auto pr-1">
      {detachments.map((d) => {
        const active = d.id === selectedId;
        return (
          <button
            key={d.id}
            type="button"
            onClick={() => onSelect(d)}
            className={cn(
              "w-full rounded-sm border p-3 text-left text-sm transition-colors",
              active
                ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
            )}
          >
            {d.name}
          </button>
        );
      })}
    </div>
  );
}
