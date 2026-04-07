import { useState } from "react";
import { ActiveSecondary, ScoringOption } from "../../types/game";

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
  currentCP: number;
  canGainCP: boolean;
  onAchieve: (secondaryId: string, vpScored: number) => void;
  onDiscard: (secondaryId: string, free: boolean) => void;
  onNewOrders: (discardSecondaryId: string) => void;
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
  currentCP,
  canGainCP,
  onAchieve,
  onDiscard,
  onNewOrders,
  onDraw,
  onScoreFixedVP,
}: Props) {
  const [expanded, setExpanded] = useState(true);

  if (!mode) return null;

  return (
    <section>
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
      >
        <span className="font-semibold">
          Secondary Missions ({mode === "tactical" ? "Tactical" : "Fixed"})
        </span>
        <span className="text-gray-400">{expanded ? "\u25B2" : "\u25BC"}</span>
      </button>

      {expanded && (
        <div className="mt-2 space-y-3">
          {/* Active Secondaries */}
          {activeSecondaries.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-xs font-semibold text-gray-400 uppercase">Active</h3>
              {activeSecondaries.map((s) => (
                <div key={s.id} className="bg-gray-800 rounded-lg p-3 border border-gray-700">
                  <div className="flex justify-between items-start mb-2">
                    <span className="font-medium text-sm">{s.name}</span>
                    <span className="text-xs text-gray-400">{s.maxVp} VP max</span>
                  </div>
                  <p className="text-xs text-gray-400 mb-3 line-clamp-2">{s.description}</p>

                  {mode === "tactical" ? (
                    <div className="flex flex-wrap gap-2">
                      {filterOptions(s.scoringOptions, "tactical").map((opt, i) => (
                        <button
                          key={i}
                          onClick={() => onAchieve(s.id, opt.vp)}
                          className="bg-green-700 hover:bg-green-600 text-white text-xs px-3 py-1 rounded transition-colors"
                          title={opt.label}
                        >
                          {opt.label} +{opt.vp}VP
                        </button>
                      ))}
                      <button
                        onClick={() => onDiscard(s.id, true)}
                        className="bg-gray-700 hover:bg-gray-600 text-gray-300 text-xs px-3 py-1 rounded transition-colors"
                      >
                        Discard
                      </button>
                      {currentRound < 5 && (
                        <button
                          onClick={() => onDiscard(s.id, false)}
                          disabled={!canGainCP}
                          className="bg-teal-800 hover:bg-teal-700 disabled:opacity-50 text-white text-xs px-3 py-1 rounded transition-colors"
                          title={
                            canGainCP
                              ? "End-of-turn discard: gain 1 CP"
                              : "CP gain cap reached this battle round"
                          }
                        >
                          {canGainCP ? "Discard +1CP" : "Discard (CP capped)"}
                        </button>
                      )}
                      <button
                        onClick={() => onNewOrders(s.id)}
                        disabled={currentCP < 1}
                        className="bg-amber-800 hover:bg-amber-700 disabled:opacity-50 text-white text-xs px-3 py-1 rounded transition-colors"
                        title="Spend 1 CP to discard and draw a new secondary"
                      >
                        New Orders
                      </button>
                    </div>
                  ) : (
                    <div className="flex flex-wrap gap-2">
                      {filterOptions(s.scoringOptions, "fixed").map((opt, i) => (
                        <button
                          key={i}
                          onClick={() => onScoreFixedVP(opt.vp)}
                          className="bg-green-700 hover:bg-green-600 text-white text-xs px-3 py-1 rounded transition-colors"
                          title={opt.label}
                        >
                          {opt.label} +{opt.vp}VP
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Tactical-mode draw button */}
          {mode === "tactical" && activeSecondaries.length < 2 && deckSize > 0 && (
            <button
              onClick={onDraw}
              className="w-full bg-indigo-600 hover:bg-indigo-700 text-white text-sm py-2 rounded-lg transition-colors"
            >
              Draw Secondaries ({deckSize} remaining)
            </button>
          )}

          {/* Deck info for tactical */}
          {mode === "tactical" && (
            <div className="text-xs text-gray-500">
              Deck: {deckSize} | Achieved: {achievedSecondaries.length} | Discarded:{" "}
              {discardedSecondaries.length}
            </div>
          )}

          {/* Achieved list */}
          {achievedSecondaries.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold text-gray-400 uppercase mb-1">Achieved</h3>
              <div className="space-y-1">
                {achievedSecondaries.map((s, i) => (
                  <div
                    key={`${s.id}-${i}`}
                    className="text-xs text-green-400 bg-green-900/20 rounded px-2 py-1"
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
