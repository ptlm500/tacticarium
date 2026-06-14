import { useState } from "react";
import { Minus, Plus } from "lucide-react";
import type { ScorePrompt } from "../../types/game";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface Props {
  prompts: ScorePrompt[];
  onConfirm: (promptId: string, count: number) => void;
}

/**
 * Renders the outstanding Layer-2 scoring prompts the engine raised for the
 * local player. Each prompt names the card and award, and the player supplies
 * the single off-board count the engine needs (e.g. "how many enemy units were
 * destroyed?"). Layer-1 awards are auto-scored and never appear here.
 */
export function ScorePromptsPanel({ prompts, onConfirm }: Props) {
  if (prompts.length === 0) return null;

  return (
    <section
      className="space-y-2 rounded-sm border border-amber-500/50 bg-amber-500/10 p-3"
      aria-label="Scoring prompts"
    >
      <h2 className="flex items-center gap-2 font-mono text-sm uppercase tracking-widest text-amber-400">
        Confirm Scoring
        <Badge
          variant="outline"
          className="border-amber-500/50 font-mono text-[10px] uppercase tracking-widest text-amber-300"
        >
          {prompts.length}
        </Badge>
      </h2>
      <p className="text-xs text-foreground/80">
        The engine needs a count it can't read off the board. Confirm each to apply its VP.
      </p>
      <div className="space-y-2">
        {prompts.map((prompt) => (
          <PromptCard key={prompt.id} prompt={prompt} onConfirm={onConfirm} />
        ))}
      </div>
    </section>
  );
}

function PromptCard({
  prompt,
  onConfirm,
}: {
  prompt: ScorePrompt;
  onConfirm: (promptId: string, count: number) => void;
}) {
  const [count, setCount] = useState(0);

  return (
    <div className="rounded-sm border border-border/60 bg-background/50 p-3">
      <div className="mb-1 flex items-center justify-between gap-2">
        <span className="text-sm font-medium text-foreground">{prompt.cardName}</span>
        <Badge
          variant="outline"
          className={cn(
            "font-mono text-[10px] uppercase tracking-widest",
            prompt.category === "primary"
              ? "border-primary/40 text-primary"
              : "border-emerald-500/40 text-emerald-400",
          )}
        >
          {prompt.category}
        </Badge>
      </div>
      {prompt.text && <p className="mb-3 text-xs text-muted-foreground">{prompt.text}</p>}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Button
            type="button"
            size="icon"
            variant="outline"
            aria-label="Decrease count"
            disabled={count <= 0}
            onClick={() => setCount((c) => Math.max(0, c - 1))}
          >
            <Minus className="size-4" />
          </Button>
          <span className="w-8 text-center font-mono text-lg tabular-nums text-foreground">
            {count}
          </span>
          <Button
            type="button"
            size="icon"
            variant="outline"
            aria-label="Increase count"
            onClick={() => setCount((c) => c + 1)}
          >
            <Plus className="size-4" />
          </Button>
        </div>
        <Button
          type="button"
          onClick={() => onConfirm(prompt.id, count)}
          className="font-mono uppercase tracking-widest"
        >
          Confirm
        </Button>
      </div>
    </div>
  );
}
