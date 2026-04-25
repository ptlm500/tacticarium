import { ScoringAction } from "../../types/mission";
import { PrimaryScoringSlot } from "../../types/scoring";
import { Button } from "@/components/ui/button";

interface Props {
  scoringRules: ScoringAction[];
  currentRound: number;
  missionScoringTiming: string;
  onScore: (vp: number, scoringSlot: PrimaryScoringSlot, ruleLabel: string) => void;
}

function resolveSlot(actionTiming: string | undefined, missionTiming: string): PrimaryScoringSlot {
  const timing = actionTiming || missionTiming;
  if (
    timing === "end_of_command_phase" ||
    timing === "end_of_battle_round" ||
    timing === "end_of_turn"
  ) {
    return timing;
  }
  return "end_of_command_phase";
}

export function MissionScoring({
  scoringRules,
  currentRound,
  missionScoringTiming,
  onScore,
}: Props) {
  if (scoringRules.length === 0) return null;

  return (
    <div className="space-y-2">
      <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Quick Score
      </h3>
      <div className="flex flex-wrap gap-2">
        {scoringRules.map((action: ScoringAction, i: number) => {
          const locked = action.minRound != null && currentRound < action.minRound;
          const slot = resolveSlot(action.scoringTiming, missionScoringTiming);
          return (
            <Button
              key={i}
              type="button"
              size="sm"
              variant="outline"
              onClick={() => onScore(action.vp, slot, action.label)}
              disabled={locked}
              title={
                locked
                  ? `Available from round ${action.minRound}`
                  : action.description || `Score ${action.vp} VP`
              }
            >
              {action.label} (+{action.vp})
              {locked && (
                <span className="ml-1 font-mono text-[10px] text-amber-400">
                  R{action.minRound}+
                </span>
              )}
            </Button>
          );
        })}
      </div>
    </div>
  );
}
