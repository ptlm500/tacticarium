import { Dices } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface Props {
  myPlayerNumber: number;
  myUsername: string;
  opponentUsername?: string;
  selected: number;
  onSelect: (playerNumber: 1 | 2) => void;
  onRandom: () => void;
}

export function FirstPlayerPicker({
  myPlayerNumber,
  myUsername,
  opponentUsername,
  selected,
  onSelect,
  onRandom,
}: Props) {
  const opponentPlayerNumber = (myPlayerNumber === 1 ? 2 : 1) as 1 | 2;
  const selectedMe = selected === myPlayerNumber;
  const selectedOpponent = selected !== 0 && selected === opponentPlayerNumber;

  const choiceClass = (active: boolean) =>
    cn(
      "flex-1 rounded-sm border px-3 py-2 text-sm font-medium transition-colors",
      active
        ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
        : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
      "disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:border-border/60 disabled:hover:bg-background/40",
    );

  return (
    <div className="space-y-2">
      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => onSelect(myPlayerNumber as 1 | 2)}
          className={choiceClass(selectedMe)}
        >
          {myUsername} (you)
        </button>
        <button
          type="button"
          onClick={() => onSelect(opponentPlayerNumber)}
          disabled={!opponentUsername}
          className={choiceClass(selectedOpponent)}
        >
          {opponentUsername ?? "Opponent"}
        </button>
      </div>
      <Button
        type="button"
        variant="outline"
        onClick={onRandom}
        disabled={!opponentUsername}
        className="w-full gap-2 font-mono uppercase tracking-widest"
      >
        <Dices className="size-4" />
        Random
      </Button>
      {selected === 0 && (
        <p className="font-mono text-[10px] uppercase tracking-widest text-amber-400">
          Pick who goes first before readying up.
        </p>
      )}
    </div>
  );
}
