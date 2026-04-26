import { Check, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface Props {
  painted: boolean;
  onToggle: (painted: boolean) => void;
}

export function ArmyPaintedToggle({ painted, onToggle }: Props) {
  return (
    <div className="space-y-2">
      <button
        type="button"
        role="switch"
        aria-checked={painted}
        onClick={() => onToggle(!painted)}
        className={cn(
          "flex w-full items-center justify-between rounded-sm border px-3 py-2 text-sm font-medium transition-colors",
          painted
            ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
            : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
        )}
      >
        <span className="flex items-center gap-2">
          {painted ? <Check className="size-4" /> : <X className="size-4 opacity-60" />}
          {painted ? "Painted" : "Not painted"}
        </span>
        <span className="font-mono text-[10px] uppercase tracking-widest">
          {painted ? "+10 VP" : "0 VP"}
        </span>
      </button>
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Hobby bonus — locked once the game starts.
      </p>
    </div>
  );
}
