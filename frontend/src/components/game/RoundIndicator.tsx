import { cn } from "@/lib/utils";

interface Props {
  currentRound: number;
  currentTurn: number;
  maxRounds: number;
}

export function RoundIndicator({ currentRound, currentTurn, maxRounds }: Props) {
  return (
    <div className="flex flex-col items-center gap-1">
      <div className="flex items-center justify-center gap-2">
        {Array.from({ length: maxRounds }, (_, i) => {
          const round = i + 1;
          const isActive = round === currentRound;
          const isPast = round < currentRound;
          return (
            <div
              key={round}
              className={cn(
                "flex h-8 w-8 items-center justify-center rounded-full border font-mono text-sm font-semibold transition-colors",
                isActive
                  ? "border-primary bg-primary/20 text-primary shadow-[0_0_8px_var(--primary)]"
                  : isPast
                    ? "border-primary/40 bg-primary/5 text-primary/70"
                    : "border-border/60 bg-background/40 text-muted-foreground",
              )}
            >
              {round}
            </div>
          );
        })}
      </div>
      <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Round {currentRound} · Turn {currentTurn} of 2
      </span>
    </div>
  );
}
