import { ScoringAction } from '../../types/mission';
import { ActiveSecondary } from '../../types/game';

export type ScoringPromptItem =
  | { kind: 'primary'; missionName: string; scoringRules: ScoringAction[]; currentRound: number }
  | { kind: 'secondary' }
  | { kind: 'fixed_secondary'; secondaries: ActiveSecondary[] }
  | { kind: 'tactical_draw' }
  | { kind: 'end_of_round_primary'; missionName: string; note: string };

interface Props {
  items: ScoringPromptItem[];
  onScore: (category: string, delta: number) => void;
  activeSecondaries: ActiveSecondary[];
  onAchieveSecondary: (id: string, vp: number) => void;
  onDiscardSecondary: (id: string, free: boolean) => void;
  onDrawSecondary: () => void;
  canGainCP: boolean;
  deckSize: number;
  activeSecondaryCount: number;
  onScoreFixedVP: (delta: number) => void;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ScoringPrompt({
  items,
  onScore,
  activeSecondaries,
  onAchieveSecondary,
  onDiscardSecondary,
  onDrawSecondary,
  canGainCP,
  deckSize,
  activeSecondaryCount,
  onScoreFixedVP,
  onConfirm,
  onCancel,
}: Props) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
      <div className="bg-gray-800 border border-gray-600 rounded-xl max-w-lg w-full mx-4 max-h-[80vh] overflow-auto">
        <div className="px-5 py-4 border-b border-gray-700">
          <h2 className="text-lg font-bold text-white">Scoring Reminder</h2>
          <p className="text-sm text-gray-400 mt-1">
            Before advancing, check if you need to score.
          </p>
        </div>

        <div className="px-5 py-4 space-y-4">
          {items.map((item, i) => (
            <div key={i}>
              {item.kind === 'primary' && (
                <PrimaryReminder
                  missionName={item.missionName}
                  scoringRules={item.scoringRules}
                  currentRound={item.currentRound}
                  onScore={(vp) => onScore('primary', vp)}
                />
              )}
              {item.kind === 'end_of_round_primary' && (
                <div className="bg-indigo-900/40 border border-indigo-700 rounded-lg p-3">
                  <h3 className="text-sm font-semibold text-indigo-200">
                    Primary Mission — {item.missionName}
                  </h3>
                  <p className="text-xs text-indigo-300 mt-1">{item.note}</p>
                </div>
              )}
              {item.kind === 'fixed_secondary' && (
                <FixedSecondaryReminder
                  secondaries={item.secondaries}
                  onScore={onScoreFixedVP}
                />
              )}
              {item.kind === 'secondary' && (
                <SecondaryReminder
                  activeSecondaries={activeSecondaries}
                  onAchieve={onAchieveSecondary}
                  onDiscard={onDiscardSecondary}
                  canGainCP={canGainCP}
                />
              )}
              {item.kind === 'tactical_draw' && (
                <TacticalDrawReminder
                  deckSize={deckSize}
                  activeCount={activeSecondaryCount}
                  onDraw={onDrawSecondary}
                />
              )}
            </div>
          ))}
        </div>

        <div className="px-5 py-4 border-t border-gray-700 flex gap-3">
          <button
            onClick={onCancel}
            className="flex-1 bg-gray-700 hover:bg-gray-600 text-white font-semibold py-2.5 rounded-lg transition-colors"
          >
            Let me score first
          </button>
          <button
            onClick={onConfirm}
            className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2.5 rounded-lg transition-colors"
          >
            I've scored, continue
          </button>
        </div>
      </div>
    </div>
  );
}

function PrimaryReminder({
  missionName,
  scoringRules,
  currentRound,
  onScore,
}: {
  missionName: string;
  scoringRules: ScoringAction[];
  currentRound: number;
  onScore: (vp: number) => void;
}) {
  return (
    <div className="bg-indigo-900/40 border border-indigo-700 rounded-lg p-3">
      <h3 className="text-sm font-semibold text-indigo-200">
        Score Primary — {missionName}
      </h3>
      <div className="flex flex-wrap gap-2 mt-2">
        {scoringRules.map((action, i) => {
          const locked = action.minRound != null && currentRound < action.minRound;
          return (
            <button
              key={i}
              onClick={() => onScore(action.vp)}
              disabled={locked}
              className="bg-indigo-800 hover:bg-indigo-700 disabled:opacity-40 disabled:cursor-not-allowed text-white text-xs px-3 py-2 rounded transition-colors"
              title={
                locked
                  ? `Available from round ${action.minRound}`
                  : action.description || `Score ${action.vp} VP`
              }
            >
              {action.label} (+{action.vp})
              {locked && <span className="ml-1 text-yellow-400">R{action.minRound}+</span>}
            </button>
          );
        })}
      </div>
    </div>
  );
}

function SecondaryReminder({
  activeSecondaries,
  onAchieve,
  onDiscard,
  canGainCP,
}: {
  activeSecondaries: ActiveSecondary[];
  onAchieve: (id: string, vp: number) => void;
  onDiscard: (id: string, free: boolean) => void;
  canGainCP: boolean;
}) {
  return (
    <div className="bg-emerald-900/40 border border-emerald-700 rounded-lg p-3">
      <h3 className="text-sm font-semibold text-emerald-200">
        Score / Discard Secondaries
      </h3>
      {activeSecondaries.length === 0 ? (
        <p className="text-xs text-emerald-300 mt-1">No active secondary missions.</p>
      ) : (
        <div className="space-y-3 mt-2">
          {activeSecondaries.map((s) => {
            const opts = (s.scoringOptions ?? []).filter(
              (o) => !o.mode || o.mode === 'tactical'
            );
            return (
              <div key={s.id}>
                <span className="text-xs text-white font-medium">{s.name}</span>
                <div className="flex flex-wrap gap-1 mt-1">
                  {opts.map((opt, i) => (
                    <button
                      key={i}
                      onClick={() => onAchieve(s.id, opt.vp)}
                      className="bg-green-700 hover:bg-green-600 text-white text-xs px-2 py-1 rounded transition-colors"
                      title={opt.label}
                    >
                      {opt.label} +{opt.vp}
                    </button>
                  ))}
                  <button
                    onClick={() => onDiscard(s.id, true)}
                    className="bg-gray-700 hover:bg-gray-600 text-gray-300 text-xs px-2 py-1 rounded transition-colors"
                  >
                    Discard
                  </button>
                  {canGainCP && (
                    <button
                      onClick={() => onDiscard(s.id, false)}
                      className="bg-teal-800 hover:bg-teal-700 text-white text-xs px-2 py-1 rounded transition-colors"
                    >
                      +1CP
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function TacticalDrawReminder({
  deckSize,
  activeCount,
  onDraw,
}: {
  deckSize: number;
  activeCount: number;
  onDraw: () => void;
}) {
  const canDraw = activeCount < 2 && deckSize > 0;
  return (
    <div className="bg-amber-900/40 border border-amber-700 rounded-lg p-3">
      <h3 className="text-sm font-semibold text-amber-200">
        Draw Tactical Secondaries
      </h3>
      <p className="text-xs text-amber-300 mt-1">
        {canDraw
          ? `You have ${activeCount} active secondary mission${activeCount === 1 ? '' : 's'}. Draw up to 2.`
          : activeCount >= 2
          ? 'You already have 2 active secondaries.'
          : 'Deck is empty.'}
      </p>
      {canDraw && (
        <button
          onClick={onDraw}
          className="mt-2 bg-amber-700 hover:bg-amber-600 text-white text-xs px-3 py-2 rounded transition-colors"
        >
          Draw Secondaries ({deckSize} remaining)
        </button>
      )}
    </div>
  );
}

function FixedSecondaryReminder({
  secondaries,
  onScore,
}: {
  secondaries: ActiveSecondary[];
  onScore: (vp: number) => void;
}) {
  return (
    <div className="bg-emerald-900/40 border border-emerald-700 rounded-lg p-3">
      <h3 className="text-sm font-semibold text-emerald-200">
        Score Fixed Secondaries
      </h3>
      <div className="space-y-3 mt-2">
        {secondaries.map((s) => {
          const opts = (s.scoringOptions ?? []).filter(
            (o) => !o.mode || o.mode === 'fixed'
          );
          return (
            <div key={s.id}>
              <p className="text-xs text-white font-medium">{s.name}</p>
              <div className="flex flex-wrap gap-1 mt-1">
                {opts.map((opt, i) => (
                  <button
                    key={i}
                    onClick={() => onScore(opt.vp)}
                    className="bg-green-700 hover:bg-green-600 text-white text-xs px-2 py-1 rounded transition-colors"
                    title={opt.label}
                  >
                    {opt.label} +{opt.vp}VP
                  </button>
                ))}
              </div>
              <p className="text-xs text-gray-400 mt-1">max {s.maxVp} VP total</p>
            </div>
          );
        })}
      </div>
    </div>
  );
}
