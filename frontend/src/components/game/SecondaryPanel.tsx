import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { ActiveSecondary } from "../../types/game";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { SecondaryDetailsModal } from "./SecondaryDetailsModal";
import { SecondaryKanbanBoard } from "./SecondaryKanbanBoard";
import { Pile, filterOptions } from "./secondaryPiles";

interface Props {
  mode: string;
  activeSecondaries: ActiveSecondary[];
  achievedSecondaries: ActiveSecondary[];
  discardedSecondaries: ActiveSecondary[];
  tacticalDeck: ActiveSecondary[];
  currentRound: number;
  currentPhase: string;
  isMyTurn: boolean;
  currentCP: number;
  canGainCP: boolean;
  newOrdersUsedThisPhase: boolean;
  onAchieve: (secondaryId: string, vpScored: number) => void;
  onDiscard: (secondaryId: string, free: boolean) => void;
  onNewOrders: (discardSecondaryId: string) => void;
  onReshuffle: (secondaryId: string) => void;
  onDraw: () => void;
  onMove: (secondaryId: string, fromPile: Pile, toPile: Pile, vpScored?: number) => void;
  onScoreFixedVP: (delta: number) => void;
}

export function SecondaryPanel({
  mode,
  activeSecondaries,
  achievedSecondaries,
  discardedSecondaries,
  tacticalDeck,
  currentRound,
  currentPhase,
  isMyTurn,
  currentCP,
  canGainCP,
  newOrdersUsedThisPhase,
  onAchieve,
  onDiscard,
  onNewOrders,
  onReshuffle,
  onDraw,
  onMove,
  onScoreFixedVP,
}: Props) {
  const showNewOrders = isMyTurn && currentPhase === "command";
  const showCPDiscard = isMyTurn && currentPhase === "fight";
  const canDraw = isMyTurn && currentPhase === "command";
  const [expanded, setExpanded] = useState(true);
  const [manageManually, setManageManually] = useState(false);
  const [detailsCard, setDetailsCard] = useState<ActiveSecondary | null>(null);
  const deckSize = tacticalDeck.length;

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
          {!showManual && activeSecondaries.length > 0 && (
            <div className="space-y-2">
              <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Active
              </h3>
              {activeSecondaries.map((s) => (
                <div key={s.id} className="rounded-sm border border-border/60 bg-background/40 p-3">
                  <button
                    type="button"
                    onClick={() => setDetailsCard(s)}
                    className="block w-full cursor-pointer text-left transition-colors hover:opacity-80"
                    title="View full details"
                  >
                    <div className="mb-2 flex items-start justify-between gap-2">
                      <span className="text-sm font-medium text-foreground">{s.name}</span>
                      <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                        {s.maxVp} VP max
                      </span>
                    </div>
                    <p className="mb-3 line-clamp-2 text-xs text-muted-foreground">
                      {s.description}
                    </p>
                  </button>

                  {isTactical ? (
                    <div className="flex flex-wrap gap-2">
                      {filterOptions(s.scoringOptions, "tactical").map((opt, i) => (
                        <Button
                          key={i}
                          type="button"
                          size="sm"
                          onClick={() => onAchieve(s.id, opt.vp)}
                          title={opt.label}
                          className="bg-emerald-600 hover:bg-emerald-700 text-white"
                        >
                          {opt.label} +{opt.vp}VP
                        </Button>
                      ))}
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => onDiscard(s.id, true)}
                      >
                        Discard
                      </Button>
                      {showCPDiscard && currentRound < 5 && (
                        <Button
                          type="button"
                          size="sm"
                          onClick={() => onDiscard(s.id, false)}
                          disabled={!canGainCP}
                          className="bg-teal-700 text-white hover:bg-teal-800"
                          title={
                            canGainCP
                              ? "End-of-turn discard: gain 1 CP"
                              : "CP gain cap reached this battle round"
                          }
                        >
                          {canGainCP ? "Discard +1CP" : "Discard (CP capped)"}
                        </Button>
                      )}
                      {showNewOrders && (
                        <Button
                          type="button"
                          size="sm"
                          onClick={() => onNewOrders(s.id)}
                          disabled={currentCP < 1 || newOrdersUsedThisPhase}
                          className="bg-amber-700 text-white hover:bg-amber-800"
                          title={
                            newOrdersUsedThisPhase
                              ? "Already used this Command phase"
                              : "Spend 1 CP to discard and draw a new secondary"
                          }
                        >
                          New Orders
                        </Button>
                      )}
                      {s.drawRestriction &&
                        s.drawRestriction.mode === "optional" &&
                        s.drawRestriction.round === currentRound && (
                          <Button
                            type="button"
                            size="sm"
                            variant="secondary"
                            onClick={() => onReshuffle(s.id)}
                            title="When Drawn: shuffle this card back into your deck and draw a replacement"
                          >
                            Shuffle Back
                          </Button>
                        )}
                    </div>
                  ) : (
                    <div className="flex flex-wrap gap-2">
                      {filterOptions(s.scoringOptions, "fixed").map((opt, i) => (
                        <Button
                          key={i}
                          type="button"
                          size="sm"
                          onClick={() => onScoreFixedVP(opt.vp)}
                          title={opt.label}
                          className="bg-emerald-600 hover:bg-emerald-700 text-white"
                        >
                          {opt.label} +{opt.vp}VP
                        </Button>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {isTactical && !showManual && activeSecondaries.length < 2 && deckSize > 0 && (
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
              activeSecondaries={activeSecondaries}
              achievedSecondaries={achievedSecondaries}
              discardedSecondaries={discardedSecondaries}
              tacticalDeck={tacticalDeck}
              onMove={onMove}
              onSelect={setDetailsCard}
            />
          )}

          {isTactical && !showManual && (
            <div className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Deck: {deckSize} | Achieved: {achievedSecondaries.length} | Discarded:{" "}
              {discardedSecondaries.length}
            </div>
          )}

          {!showManual && achievedSecondaries.length > 0 && (
            <div>
              <h3 className="mb-1 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Achieved
              </h3>
              <div className="space-y-1">
                {achievedSecondaries.map((s, i) => (
                  <button
                    type="button"
                    key={`${s.id}-${i}`}
                    onClick={() => setDetailsCard(s)}
                    className={cn(
                      "w-full rounded-sm border border-emerald-500/40 bg-emerald-500/10 px-2 py-1 text-left transition-colors hover:bg-emerald-500/20",
                      "text-xs text-emerald-400",
                    )}
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
