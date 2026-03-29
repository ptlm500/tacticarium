import { ScoringAction } from "../../types/mission";

interface Props {
  scoringRules: ScoringAction[];
  currentRound: number;
  onScore: (vp: number) => void;
}

export function MissionScoring({ scoringRules, currentRound, onScore }: Props) {
  if (scoringRules.length === 0) return null;

  return (
    <div className="space-y-2">
      <h3 className="text-xs font-semibold text-gray-400 uppercase">Quick Score</h3>
      <div className="flex flex-wrap gap-2">
        {scoringRules.map((action: ScoringAction, i: number) => {
          const locked = action.minRound != null && currentRound < action.minRound;
          return (
            <button
              key={i}
              onClick={() => onScore(action.vp)}
              disabled={locked}
              className="bg-gray-700 hover:bg-gray-600 disabled:opacity-40 disabled:cursor-not-allowed text-white text-xs px-3 py-2 rounded transition-colors"
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
