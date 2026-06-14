import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { SecondaryCard } from "../../types/game";
import { Button } from "@/components/ui/button";
import { SecondaryDetailsModal } from "./SecondaryDetailsModal";
import { SecondaryKanbanBoard } from "./SecondaryKanbanBoard";
import { Pile } from "./secondaryPiles";

interface Props {
  mode: string;
  secondaryHand: SecondaryCard[];
  secondaryScored: SecondaryCard[];
  secondaryDeck: SecondaryCard[];
  currentPhase: string;
  isMyTurn: boolean;
  onDiscard: (secondaryId: string, free: boolean) => void;
  onDraw: () => void;
  onMove: (secondaryId: string, fromPile: Pile, toPile: Pile, vpScored?: number) => void;
}

export function SecondaryPanel({
  mode,
  secondaryHand,
  secondaryScored,
  secondaryDeck,
  currentPhase,
  isMyTurn,
  onDiscard,
  onDraw,
  onMove,
}: Props) {
  const canDraw = isMyTurn && currentPhase === "command";
  const [expanded, setExpanded] = useState(true);
  const [manageManually, setManageManually] = useState(false);
  const [detailsCard, setDetailsCard] = useState<SecondaryCard | null>(null);
  const deckSize = secondaryDeck.length;

  if (!mode) return null;

  const isTactical = mode === "tactical";
  const showManual = isTactical && manageManually;

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
          {isTactical && (
            <label className="flex items-center gap-2 rounded-sm border border-border/60 bg-background/40 px-3 py-2 text-xs text-muted-foreground">
              <input
                type="checkbox"
                checked={manageManually}
                onChange={(e) => setManageManually(e.target.checked)}
              />
              <span className="font-mono uppercase tracking-widest">Manage manually</span>
            </label>
          )}

          {/* Active pile (only in non-manual mode — the kanban board renders all piles below) */}
          {!showManual && secondaryHand.length > 0 && (
            <div className="space-y-2">
              <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Active
              </h3>
              {secondaryHand.map((s) => (
                <div key={s.id} className="rounded-sm border border-border/60 bg-background/40 p-3">
                  <button
                    type="button"
                    onClick={() => setDetailsCard(s)}
                    className="block w-full cursor-pointer text-left transition-colors hover:opacity-80"
                    title="View full details"
                  >
                    <div className="mb-2 flex items-start justify-between gap-2">
                      <span className="text-sm font-medium text-foreground">{s.name}</span>
                    </div>
                    <p className="mb-3 line-clamp-2 text-xs text-muted-foreground">{s.text}</p>
                  </button>

                  {isTactical && (
                    <div className="flex flex-wrap gap-2">
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => onDiscard(s.id, true)}
                      >
                        Discard
                      </Button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {isTactical && !showManual && secondaryHand.length < 2 && deckSize > 0 && (
            <Button
              type="button"
              onClick={onDraw}
              disabled={!canDraw}
              title={
                canDraw
                  ? undefined
                  : !isMyTurn
                    ? "Only the active player can draw secondaries"
                    : "Drawing is restricted to the Command phase"
              }
              className="w-full font-mono uppercase tracking-widest"
            >
              Draw Secondaries ({deckSize} remaining)
            </Button>
          )}

          {/* Manual kanban board — drag cards between piles, mirrors the physical deck */}
          {showManual && (
            <SecondaryKanbanBoard
              activeSecondaries={secondaryHand}
              achievedSecondaries={secondaryScored}
              discardedSecondaries={[]}
              tacticalDeck={secondaryDeck}
              onMove={onMove}
              onSelect={setDetailsCard}
            />
          )}

          {isTactical && !showManual && (
            <div className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Deck: {deckSize} | Scored: {secondaryScored.length}
            </div>
          )}

          {!showManual && secondaryScored.length > 0 && (
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
                    className="w-full rounded-sm border border-emerald-500/40 bg-emerald-500/10 px-2 py-1 text-left text-xs text-emerald-400 transition-colors hover:bg-emerald-500/20"
                    title="View full details"
                  >
                    {s.name}
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
