import { useState } from "react";
import { Minus, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";

interface Props {
  vpPrimary: number;
  vpSecondary: number;
  vpGambit: number;
  vpPaint: number;
  onAdjust: (category: string, delta: number) => void;
}

export function VPCounter({ vpPrimary, vpSecondary, vpGambit, vpPaint, onAdjust }: Props) {
  const [expanded, setExpanded] = useState(false);
  const total = vpPrimary + vpSecondary + vpGambit + vpPaint;

  return (
    <div className="text-center">
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Victory Points
      </p>
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="font-mono text-3xl font-bold text-primary tabular-nums hover:text-primary/80"
      >
        {total}
      </button>

      {expanded && (
        <div className="mt-3 space-y-2 text-sm">
          <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            Manual adjust (bypasses mission rules)
          </p>
          <VPRow
            label="Primary"
            value={vpPrimary}
            max={50}
            category="primary"
            onAdjust={onAdjust}
          />
          <VPRow
            label="Secondary"
            value={vpSecondary}
            max={40}
            category="secondary"
            onAdjust={onAdjust}
          />
          <VPRow label="Gambit" value={vpGambit} max={12} category="gambit" onAdjust={onAdjust} />
          <div className="flex items-center justify-between text-muted-foreground">
            <span>Paint</span>
            <span className="tabular-nums">{vpPaint}/10</span>
          </div>
        </div>
      )}
    </div>
  );
}

function VPRow({
  label,
  value,
  max,
  category,
  onAdjust,
}: {
  label: string;
  value: number;
  max: number;
  category: string;
  onAdjust: (category: string, delta: number) => void;
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant="outline"
          size="icon-sm"
          onClick={() => onAdjust(category, -1)}
          disabled={value <= 0}
          aria-label={`Decrease ${label} VP`}
        >
          <Minus className="size-3" />
        </Button>
        <span className="w-12 text-center font-mono tabular-nums">
          {value}/{max}
        </span>
        <Button
          type="button"
          variant="outline"
          size="icon-sm"
          onClick={() => onAdjust(category, 1)}
          disabled={value >= max}
          aria-label={`Increase ${label} VP`}
        >
          <Plus className="size-3" />
        </Button>
      </div>
    </div>
  );
}
