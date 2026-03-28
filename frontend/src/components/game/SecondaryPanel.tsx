import { useState } from 'react';
import { ActiveSecondary } from '../../types/game';

interface Props {
  mode: string;
  activeSecondaries: ActiveSecondary[];
  achievedSecondaries: ActiveSecondary[];
  discardedSecondaries: ActiveSecondary[];
  deckSize: number;
  currentRound: number;
  currentCP: number;
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
  onAchieve,
  onDiscard,
  onNewOrders,
  onDraw,
  onScoreFixedVP,
}: Props) {
  const [expanded, setExpanded] = useState(true);
  const [achieveVP, setAchieveVP] = useState<Record<string, number>>({});

  if (!mode) return null;

  return (
    <section>
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
      >
        <span className="font-semibold">
          Secondary Missions ({mode === 'tactical' ? 'Tactical' : 'Fixed'})
        </span>
        <span className="text-gray-400">{expanded ? '\u25B2' : '\u25BC'}</span>
      </button>

      {expanded && (
        <div className="mt-2 space-y-3">
          {/* Active Secondaries */}
          {activeSecondaries.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-xs font-semibold text-gray-400 uppercase">Active</h3>
              {activeSecondaries.map((s) => (
                <div
                  key={s.id}
                  className="bg-gray-800 rounded-lg p-3 border border-gray-700"
                >
                  <div className="flex justify-between items-start mb-2">
                    <span className="font-medium text-sm">{s.name}</span>
                    <span className="text-xs text-gray-400">{s.maxVp} VP max</span>
                  </div>
                  <p className="text-xs text-gray-400 mb-3 line-clamp-2">
                    {s.description}
                  </p>

                  {mode === 'tactical' ? (
                    <div className="flex gap-2">
                      <div className="flex items-center gap-1">
                        <input
                          type="number"
                          min={1}
                          max={s.maxVp}
                          value={achieveVP[s.id] ?? s.maxVp}
                          onChange={(e) =>
                            setAchieveVP({ ...achieveVP, [s.id]: Number(e.target.value) })
                          }
                          className="w-12 bg-gray-900 border border-gray-600 rounded px-2 py-1 text-xs text-center"
                        />
                        <button
                          onClick={() => onAchieve(s.id, achieveVP[s.id] ?? s.maxVp)}
                          className="bg-green-700 hover:bg-green-600 text-white text-xs px-3 py-1 rounded transition-colors"
                        >
                          Achieve
                        </button>
                      </div>
                      <button
                        onClick={() => onDiscard(s.id, true)}
                        className="bg-gray-700 hover:bg-gray-600 text-gray-300 text-xs px-3 py-1 rounded transition-colors"
                      >
                        Discard
                      </button>
                      {currentRound < 5 && (
                        <button
                          onClick={() => onDiscard(s.id, false)}
                          className="bg-teal-800 hover:bg-teal-700 text-white text-xs px-3 py-1 rounded transition-colors"
                          title="End-of-turn discard: gain 1 CP"
                        >
                          Discard +1CP
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
                    <div className="flex gap-2">
                      <button
                        onClick={() => onScoreFixedVP(s.maxVp)}
                        className="bg-green-700 hover:bg-green-600 text-white text-xs px-3 py-1 rounded transition-colors"
                      >
                        Score {s.maxVp} VP
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Tactical-mode draw button */}
          {mode === 'tactical' && activeSecondaries.length < 2 && deckSize > 0 && (
            <button
              onClick={onDraw}
              className="w-full bg-indigo-600 hover:bg-indigo-700 text-white text-sm py-2 rounded-lg transition-colors"
            >
              Draw Secondaries ({deckSize} remaining)
            </button>
          )}

          {/* Deck info for tactical */}
          {mode === 'tactical' && (
            <div className="text-xs text-gray-500">
              Deck: {deckSize} | Achieved: {achievedSecondaries.length} | Discarded:{' '}
              {discardedSecondaries.length}
            </div>
          )}

          {/* Achieved list */}
          {achievedSecondaries.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold text-gray-400 uppercase mb-1">
                Achieved
              </h3>
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
