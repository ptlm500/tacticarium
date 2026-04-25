import { Minus, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";

interface Props {
  cp: number;
  canGainCP: boolean;
  onAdjust: (delta: number) => void;
}

export function CPCounter({ cp, canGainCP, onAdjust }: Props) {
  return (
    <div className="text-center">
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Command Points
      </p>
      <div className="mt-1 flex items-center justify-center gap-3">
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={() => onAdjust(-1)}
          disabled={cp <= 0}
          aria-label="Decrease CP"
        >
          <Minus className="size-4" />
        </Button>
        <span className="w-12 text-center font-mono text-3xl font-bold text-primary tabular-nums">
          {cp}
        </span>
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={() => onAdjust(1)}
          aria-label="Increase CP"
          title={!canGainCP ? "CP gain cap reached — confirmation required" : "Gain 1 CP"}
        >
          <Plus className="size-4" />
        </Button>
      </div>
      {!canGainCP && (
        <p className="mt-1 font-mono text-[10px] uppercase tracking-widest text-amber-400">
          CP gain cap reached
        </p>
      )}
    </div>
  );
}
