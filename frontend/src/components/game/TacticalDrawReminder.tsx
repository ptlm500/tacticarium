import { Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";

interface Props {
  deckSize: number;
  activeCount: number;
  onDraw: () => void;
}

export function TacticalDrawReminder({ deckSize, activeCount, onDraw }: Props) {
  const canDraw = activeCount < 2 && deckSize > 0;
  return (
    <div className="rounded-sm border border-amber-500/40 bg-amber-500/10 p-3">
      <h3 className="font-mono text-sm uppercase tracking-widest text-amber-400">
        Draw Tactical Secondaries
      </h3>
      <p className="mt-1 text-xs text-foreground/80">
        {canDraw
          ? `You have ${activeCount}/2 active secondaries. Draw to fill your active slots.`
          : activeCount >= 2
            ? "You already have 2 active secondaries."
            : "Deck is empty."}
      </p>
      {canDraw && (
        <Button
          type="button"
          size="sm"
          onClick={onDraw}
          className="mt-2 gap-1 bg-amber-600 text-white hover:bg-amber-700"
        >
          <Sparkles className="size-3" />
          Draw Secondaries ({deckSize} remaining)
        </Button>
      )}
    </div>
  );
}
