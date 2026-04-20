import { Phase, PHASE_LABELS } from "../../types/game";
import { cn } from "@/lib/utils";

interface Props {
  currentPhase: Phase;
  phases: Phase[];
}

export function PhaseTracker({ currentPhase, phases }: Props) {
  return (
    <div className="flex items-center gap-1 overflow-x-auto">
      {phases.map((phase, i) => {
        const isActive = phase === currentPhase;
        const isPast = phases.indexOf(currentPhase) > i;
        return (
          <div key={phase} className="flex items-center">
            {i > 0 && (
              <div
                className={cn(
                  "h-0.5 w-4 transition-colors",
                  isPast || isActive ? "bg-primary" : "bg-border/60",
                )}
              />
            )}
            <span
              className={cn(
                "whitespace-nowrap rounded-sm border px-2 py-1 font-mono text-[10px] uppercase tracking-widest transition-colors",
                isActive
                  ? "border-primary bg-primary/20 text-primary shadow-[0_0_6px_var(--primary)]"
                  : isPast
                    ? "border-primary/40 text-primary/80"
                    : "border-border/40 text-muted-foreground",
              )}
            >
              {PHASE_LABELS[phase]}
            </span>
          </div>
        );
      })}
    </div>
  );
}
