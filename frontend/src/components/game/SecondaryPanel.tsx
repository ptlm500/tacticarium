import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { SecondaryCard } from "../../types/game";
import { Button } from "@/components/ui/button";
import { SecondaryDetailsModal } from "./SecondaryDetailsModal";

interface Props {
  mode: string;
  secondaryHand: SecondaryCard[];
  secondaryScored: SecondaryCard[];
  secondaryDeck: SecondaryCard[];
  currentPhase: string;
  isMyTurn: boolean;
  /** Draw 2 cards from the deck, keeping the current hand (tactical only). */
  onDraw: () => void;
}

export function SecondaryPanel({
  mode,
  secondaryHand,
  secondaryScored,
  secondaryDeck,
  currentPhase,
  isMyTurn,
  onDraw,
}: Props) {
  const [expanded, setExpanded] = useState(true);
  const [detailsCard, setDetailsCard] = useState<SecondaryCard | null>(null);

  if (!mode) return null;

  const isTactical = mode === "tactical";
  const deckSize = secondaryDeck.length;
  // 11e: draw 2 each Command phase, keeping all unscored cards (no hand limit).
  const canDraw = isTactical && isMyTurn && currentPhase === "command" && deckSize > 0;

  return (
    <section>
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between rounded-sm border border-border/60 bg-background/40 px-4 py-3 text-left transition-colors hover:border-primary/50"
      >
        <span className="font-mono text-sm uppercase tracking-widest text-primary">
          Secondary Missions ({isTactical ? "Tactical" : "Fixed"})
        </span>
        {expanded ? (
          <ChevronUp className="size-4 text-muted-foreground" />
        ) : (
          <ChevronDown className="size-4 text-muted-foreground" />
        )}
      </button>

      {expanded && (
        <div className="mt-2 space-y-3">
          {secondaryHand.length > 0 ? (
            <div className="space-y-2">
              <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Hand
              </h3>
              {secondaryHand.map((s) => (
                <button
                  type="button"
                  key={s.id}
                  onClick={() => setDetailsCard(s)}
                  title="View full details"
                  className="block w-full rounded-sm border border-border/60 bg-background/40 p-3 text-left transition-colors hover:border-primary/50"
                >
                  <span className="text-sm font-medium text-foreground">{s.name}</span>
                  <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{s.text}</p>
                </button>
              ))}
            </div>
          ) : (
            <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              No cards in hand
            </p>
          )}

          {isTactical && (
            <Button
              type="button"
              onClick={onDraw}
              disabled={!canDraw}
              title={
                deckSize === 0
                  ? "Deck is empty"
                  : !isMyTurn
                    ? "Only the active player can draw secondaries"
                    : currentPhase !== "command"
                      ? "Drawing is restricted to the Command phase"
                      : undefined
              }
              className="w-full font-mono uppercase tracking-widest"
            >
              Draw Secondaries ({deckSize} in deck)
            </Button>
          )}

          <div className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            Deck: {deckSize} | Scored: {secondaryScored.length}
          </div>

          {secondaryScored.length > 0 && (
            <div>
              <h3 className="mb-1 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Scored
              </h3>
              <div className="space-y-1">
                {secondaryScored.map((s, i) => (
                  <button
                    type="button"
                    key={`${s.id}-${i}`}
                    onClick={() => setDetailsCard(s)}
                    className="flex w-full items-center justify-between rounded-sm border border-emerald-500/40 bg-emerald-500/10 px-2 py-1 text-left text-xs text-emerald-400 transition-colors hover:bg-emerald-500/20"
                    title="View full details"
                  >
                    <span>{s.name}</span>
                    {s.vpScored != null && s.vpScored > 0 && (
                      <span className="font-mono tabular-nums">+{s.vpScored}</span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      <SecondaryDetailsModal secondary={detailsCard} onClose={() => setDetailsCard(null)} />
    </section>
  );
}
