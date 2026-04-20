import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { ActiveSecondary, ScoringOption } from "../../types/game";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

function filterOptions(options: ScoringOption[] | null | undefined, mode: string): ScoringOption[] {
  if (!options || options.length === 0) return [];
  return options.filter((o) => !o.mode || o.mode === mode);
}

interface Props {
  mode: string;
  activeSecondaries: ActiveSecondary[];
  achievedSecondaries: ActiveSecondary[];
  discardedSecondaries: ActiveSecondary[];
  deckSize: number;
  currentRound: number;
  currentPhase: string;
  isMyTurn: boolean;
  currentCP: number;
  canGainCP: boolean;
  onAchieve: (secondaryId: string, vpScored: number) => void;
  onDiscard: (secondaryId: string, free: boolean) => void;
  onNewOrders: (discardSecondaryId: string) => void;
  onReshuffle: (secondaryId: string) => void;
  onDraw: () => void;
  onScoreFixedVP: (delta: number) => void;
}

export function SecondaryPanel({
  mode,
  activeSecondaries,
  achievedSecondaries,
  discardedSecondaries,
  deckSize,
  currentRound,
  currentPhase,
  isMyTurn,
  currentCP,
  canGainCP,
  onAchieve,
  onDiscard,
  onNewOrders,
  onReshuffle,
  onDraw,
  onScoreFixedVP,
}: Props) {
  const showNewOrders = isMyTurn && currentPhase === "command";
  const showCPDiscard = isMyTurn && currentPhase === "fight";
  const [expanded, setExpanded] = useState(true);

  if (!mode) return null;

  return (
    <section>
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between rounded-sm border border-border/60 bg-background/40 px-4 py-3 text-left transition-colors hover:border-primary/50"
      >
        <span className="font-mono text-sm uppercase tracking-widest text-primary">
          Secondary Missions ({mode === "tactical" ? "Tactical" : "Fixed"})
        </span>
        {expanded ? (
          <ChevronUp className="size-4 text-muted-foreground" />
        ) : (
          <ChevronDown className="size-4 text-muted-foreground" />
        )}
      </button>

      {expanded && (
        <div className="mt-2 space-y-3">
          {activeSecondaries.length > 0 && (
            <div className="space-y-2">
              <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Active
              </h3>
              {activeSecondaries.map((s) => (
                <div key={s.id} className="rounded-sm border border-border/60 bg-background/40 p-3">
                  <div className="mb-2 flex items-start justify-between gap-2">
                    <span className="text-sm font-medium text-foreground">{s.name}</span>
                    <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                      {s.maxVp} VP max
                    </span>
                  </div>
                  <p className="mb-3 line-clamp-2 text-xs text-muted-foreground">{s.description}</p>

                  {mode === "tactical" ? (
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
                          disabled={currentCP < 1}
                          className="bg-amber-700 text-white hover:bg-amber-800"
                          title="Spend 1 CP to discard and draw a new secondary"
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

          {mode === "tactical" && activeSecondaries.length < 2 && deckSize > 0 && (
            <Button
              type="button"
              onClick={onDraw}
              disabled={!isMyTurn}
              title={isMyTurn ? undefined : "Only the active player can draw secondaries"}
              className="w-full font-mono uppercase tracking-widest"
            >
              Draw Secondaries ({deckSize} remaining)
            </Button>
          )}

          {mode === "tactical" && (
            <div className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Deck: {deckSize} | Achieved: {achievedSecondaries.length} | Discarded:{" "}
              {discardedSecondaries.length}
            </div>
          )}

          {achievedSecondaries.length > 0 && (
            <div>
              <h3 className="mb-1 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Achieved
              </h3>
              <div className="space-y-1">
                {achievedSecondaries.map((s, i) => (
                  <div
                    key={`${s.id}-${i}`}
                    className={cn(
                      "rounded-sm border border-emerald-500/40 bg-emerald-500/10 px-2 py-1",
                      "text-xs text-emerald-400",
                    )}
                  >
                    {s.name}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </section>
  );
}
