import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Mission, MissionRule, ScoringAction } from "../../types/mission";
import { Badge } from "@/components/ui/badge";

interface Props {
  mission: Mission | null;
  twist: MissionRule | null;
}

export function MissionInfo({ mission, twist }: Props) {
  const [expanded, setExpanded] = useState(false);

  return (
    <section>
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between rounded-sm border border-border/60 bg-background/40 px-4 py-3 text-left transition-colors hover:border-primary/50"
      >
        <span className="font-mono text-sm uppercase tracking-widest text-primary">
          Mission Info
        </span>
        {expanded ? (
          <ChevronUp className="size-4 text-muted-foreground" />
        ) : (
          <ChevronDown className="size-4 text-muted-foreground" />
        )}
      </button>
      {expanded && (
        <div className="mt-2 space-y-4 rounded-sm border border-border/60 bg-background/40 p-4 text-sm">
          <div className="space-y-2">
            <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Primary Mission
            </h3>
            {mission ? (
              <>
                <p className="font-medium text-foreground">{mission.name}</p>
                <p className="text-foreground/80">{mission.description}</p>
                {mission.scoringRules && mission.scoringRules.length > 0 && (
                  <ScoringRulesList rules={mission.scoringRules} />
                )}
              </>
            ) : (
              <p className="text-muted-foreground">None</p>
            )}
          </div>

          <div className="space-y-2 border-t border-border/40 pt-4">
            <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Twist
            </h3>
            {twist ? (
              <>
                <p className="font-medium text-foreground">{twist.name}</p>
                <p className="text-foreground/80">{twist.description}</p>
              </>
            ) : (
              <p className="text-muted-foreground">None</p>
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
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Scoring:
      </p>
      <ul className="space-y-1">
        {rules.map((rule, i) => (
          <li key={i} className="flex items-center gap-2 text-xs">
            <Badge
              variant="outline"
              className="border-primary/40 bg-primary/10 font-mono uppercase tracking-widest text-primary"
            >
              +{rule.vp} VP
            </Badge>
            <span className="text-foreground/80">{rule.label}</span>
            {rule.minRound != null && rule.minRound > 1 && (
              <span className="font-mono text-[10px] uppercase tracking-widest text-amber-400">
                (R{rule.minRound}+)
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
