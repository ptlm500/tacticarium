import { Shield, Swords } from "lucide-react";
import { cn } from "@/lib/utils";

type Side = "attacker" | "defender";

interface Props {
  selected: string;
  onSelect: (side: Side) => void;
}

const OPTIONS: { side: Side; label: string; icon: typeof Swords; hint: string }[] = [
  {
    side: "attacker",
    label: "Attacker",
    icon: Swords,
    hint: "Deploys second, often pushes the table",
  },
  { side: "defender", label: "Defender", icon: Shield, hint: "Holds territory, deploys first" },
];

export function SidePicker({ selected, onSelect }: Props) {
  return (
    <div className="grid grid-cols-2 gap-2">
      {OPTIONS.map(({ side, label, icon: Icon, hint }) => {
        const active = selected === side;
        return (
          <button
            key={side}
            type="button"
            onClick={() => onSelect(side)}
            className={cn(
              "flex flex-col items-center gap-1 rounded-sm border px-4 py-3 text-center transition-colors",
              active
                ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
            )}
          >
            <Icon className="size-5" />
            <span className="font-mono text-sm uppercase tracking-widest">{label}</span>
            <span className="text-[10px] leading-tight text-muted-foreground">{hint}</span>
          </button>
        );
      })}
    </div>
  );
}
