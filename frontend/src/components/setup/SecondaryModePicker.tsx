import { cn } from "@/lib/utils";

interface Props {
  mode: string;
  onModeChange: (mode: "fixed" | "tactical") => void;
}

const MODES: { mode: "fixed" | "tactical"; label: string; hint: string }[] = [
  {
    mode: "tactical",
    label: "Tactical",
    hint: "Draw 2 cards each Command phase and keep every unscored card — no hand limit.",
  },
  {
    mode: "fixed",
    label: "Fixed",
    hint: "Play a set hand for the whole game, scoring the fixed-mode awards.",
  },
];

export function SecondaryModePicker({ mode, onModeChange }: Props) {
  return (
    <div className="space-y-3">
      <div className="flex gap-2">
        {MODES.map(({ mode: m, label }) => (
          <button
            key={m}
            type="button"
            onClick={() => onModeChange(m)}
            className={cn(
              "flex-1 rounded-sm border px-4 py-2 font-mono text-sm uppercase tracking-widest transition-colors",
              mode === m
                ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                : "border-border/60 bg-background/40 text-muted-foreground hover:border-primary/50 hover:text-foreground",
            )}
          >
            {label}
          </button>
        ))}
      </div>
      {mode && (
        <p className="text-xs leading-snug text-muted-foreground">
          {MODES.find((m) => m.mode === mode)?.hint}
        </p>
      )}
    </div>
  );
}
