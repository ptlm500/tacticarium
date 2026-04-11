import { useState } from "react";
import { Mission, MissionRule, ScoringAction } from "../../types/mission";

interface Props {
  mission: Mission | null;
  twist: MissionRule | null;
}

export function MissionInfo({ mission, twist }: Props) {
  const [expanded, setExpanded] = useState(false);

  return (
    <section>
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
      >
        <span className="font-semibold">Mission Info</span>
        <span className="text-gray-400">{expanded ? "\u25B2" : "\u25BC"}</span>
      </button>
      {expanded && (
        <div className="mt-2 bg-gray-800 rounded-lg p-4 space-y-4 text-sm">
          {/* Primary Mission */}
          <div className="space-y-2">
            <h3 className="text-xs font-semibold text-gray-400 uppercase">Primary Mission</h3>
            {mission ? (
              <>
                <p className="text-white font-medium">{mission.name}</p>
                <p className="text-gray-300">{mission.description}</p>
                {mission.scoringRules && mission.scoringRules.length > 0 && (
                  <ScoringRulesList rules={mission.scoringRules} />
                )}
              </>
            ) : (
              <p className="text-gray-500">None</p>
            )}
          </div>

          {/* Twist */}
          <div className="space-y-2 border-t border-gray-700 pt-4">
            <h3 className="text-xs font-semibold text-gray-400 uppercase">Twist</h3>
            {twist ? (
              <>
                <p className="text-white font-medium">{twist.name}</p>
                <p className="text-gray-300">{twist.description}</p>
              </>
            ) : (
              <p className="text-gray-500">None</p>
            )}
          </div>
        </div>
      )}
    </section>
  );
}

function ScoringRulesList({ rules }: { rules: ScoringAction[] }) {
  return (
    <div className="space-y-1">
      <p className="text-xs text-gray-400 font-medium">Scoring:</p>
      <ul className="space-y-1">
        {rules.map((rule, i) => (
          <li key={i} className="flex items-center gap-2 text-xs">
            <span className="bg-indigo-900/60 text-indigo-200 px-2 py-0.5 rounded font-medium">
              +{rule.vp} VP
            </span>
            <span className="text-gray-300">{rule.label}</span>
            {rule.minRound != null && rule.minRound > 1 && (
              <span className="text-yellow-400 text-xs">(R{rule.minRound}+)</span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
