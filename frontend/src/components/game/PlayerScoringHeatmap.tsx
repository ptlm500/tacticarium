import { cn } from "@/lib/utils";
import { HUDFrame } from "@/components/ui/hud-frame";
import type { PlayerSummaryStats, VPCategory } from "./vpUtils";

interface Props {
  username: string;
  stats: PlayerSummaryStats;
  rounds: number[];
  intensityMax: number;
  onCellClick: (round: number, category: VPCategory) => void;
}

const ROW_LABELS: Record<VPCategory, string> = {
  primary: "Pri",
  secondary: "Sec",
};

const INTENSITY_STEPS: ReadonlyArray<readonly [number, string]> = [
  [0.25, "bg-primary/15"],
  [0.5, "bg-primary/30"],
  [0.75, "bg-primary/55"],
  [Infinity, "bg-primary/80 shadow-[0_0_6px_var(--primary)]"],
];

export function PlayerScoringHeatmap({
  username,
  stats,
  rounds,
  intensityMax,
  onCellClick,
}: Props) {
  return (
    <HUDFrame label={`${username} — Scoring`}>
      <div
        className="grid gap-1"
        style={{ gridTemplateColumns: `2.25rem repeat(${rounds.length}, minmax(0, 1fr))` }}
      >
        <div />
        {rounds.map((r) => (
          <div
            key={`hdr-${r}`}
            className="text-center font-mono text-[9px] uppercase tracking-widest text-foreground/40"
          >
            R{r}
          </div>
        ))}

        {(["primary", "secondary"] as VPCategory[]).map((cat) => (
          <Row
            key={cat}
            label={ROW_LABELS[cat]}
            rounds={rounds}
            getValue={(r) => stats.vpByRound[r]?.[cat] ?? 0}
            intensityMax={intensityMax}
            onCellClick={(round) => onCellClick(round, cat)}
          />
        ))}
      </div>
    </HUDFrame>
  );
}

interface RowProps {
  label: string;
  rounds: number[];
  getValue: (round: number) => number;
  intensityMax: number;
  onCellClick: (round: number) => void;
}

function Row({ label, rounds, getValue, intensityMax, onCellClick }: RowProps) {
  return (
    <>
      <div className="flex items-center justify-end pr-1 font-mono text-[9px] uppercase tracking-widest text-foreground/50">
        {label}
      </div>
      {rounds.map((r) => {
        const value = getValue(r);
        const isEmpty = value <= 0;
        return (
          <button
            key={`${label}-${r}`}
            type="button"
            disabled={isEmpty}
            onClick={() => onCellClick(r)}
            className={cn(
              "relative flex h-9 items-center justify-center rounded-sm font-mono text-xs tabular-nums transition-all",
              cellClass(value, intensityMax),
              isEmpty
                ? "cursor-default text-foreground/30"
                : "cursor-pointer text-foreground hover:ring-1 hover:ring-primary/60 hover:ring-offset-0",
            )}
            aria-label={`${label} round ${r}: ${value} VP${isEmpty ? "" : ", click for details"}`}
          >
            {isEmpty ? "—" : value}
          </button>
        );
      })}
    </>
  );
}

function cellClass(value: number, max: number): string {
  if (value <= 0) return "bg-foreground/5";
  const intensity = max > 0 ? value / max : 0;
  return INTENSITY_STEPS.find(([threshold]) => intensity < threshold)![1];
}
