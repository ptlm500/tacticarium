import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Mission, MissionRule, ScoringAction } from "../../types/mission";
import { Badge } from "@/components/ui/badge";

interface Props {
  mission: Mission | null;
  twist: MissionRule | null;
}

interface DescriptionBlock {
  heading?: string;
  body: string;
}

function splitDescription(desc: string): DescriptionBlock[] {
  const trimmed = desc.trim();
  if (!trimmed) return [];
  const blocks = trimmed
    .split(/\n\s*\n/)
    .map((b) => b.trim())
    .filter(Boolean);
  return blocks.map((block) => {
    const lines = block.split("\n");
    const first = lines[0]?.trim() ?? "";
    const isAllCapsHeading =
      /[A-Z]/.test(first) && /^[A-Z0-9][A-Z0-9 :,.'()/&-]+$/.test(first) && !first.includes(":");
    if (isAllCapsHeading && lines.length > 1) {
      return { heading: first, body: lines.slice(1).join("\n").trim() };
    }
    return { body: block };
  });
}

function groupByMinRound(
  rules: ScoringAction[],
): Array<{ minRound: number; rules: ScoringAction[] }> {
  const groups = new Map<number, ScoringAction[]>();
  for (const r of rules) {
    const key = r.minRound ?? 1;
    const existing = groups.get(key);
    if (existing) {
      existing.push(r);
    } else {
      groups.set(key, [r]);
    }
  }
  return [...groups.entries()]
    .sort(([a], [b]) => a - b)
    .map(([minRound, rules]) => ({ minRound, rules }));
}

function roundLabel(minRound: number): string {
  if (minRound <= 1) return "Anytime";
  return `From Battle Round ${minRound}`;
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
          <div className="space-y-3">
            <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Primary Mission
            </h3>
            {mission ? (
              <>
                <p className="font-medium text-foreground">{mission.name}</p>
                <DescriptionBody text={mission.description} />
                {mission.scoringRules && mission.scoringRules.length > 0 && (
                  <ScoringRulesList rules={mission.scoringRules} />
                )}
              </>
            ) : (
              <p className="text-muted-foreground">None</p>
            )}
          </div>

          <div className="space-y-3 border-t border-border/40 pt-4">
            <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Twist
            </h3>
            {twist ? (
              <>
                <p className="font-medium text-foreground">{twist.name}</p>
                <DescriptionBody text={twist.description} />
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

function DescriptionBody({ text }: { text: string }) {
  const blocks = splitDescription(text);
  if (blocks.length === 0) return null;
  if (blocks.length === 1 && !blocks[0].heading) {
    return <p className="whitespace-pre-line text-foreground/80">{blocks[0].body}</p>;
  }
  return (
    <div className="space-y-3">
      {blocks.map((block, i) => (
        <div key={i} className="space-y-1">
          {block.heading && (
            <p className="font-mono text-[10px] uppercase tracking-widest text-amber-400">
              {block.heading}
            </p>
          )}
          {block.body && <p className="whitespace-pre-line text-foreground/80">{block.body}</p>}
        </div>
      ))}
    </div>
  );
}

function ScoringRulesList({ rules }: { rules: ScoringAction[] }) {
  const groups = groupByMinRound(rules);
  return (
    <div className="space-y-2">
      <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        Scoring:
      </p>
      <div className="space-y-3">
        {groups.map(({ minRound, rules }) => (
          <div key={minRound} className="space-y-1">
            <p className="font-mono text-[10px] uppercase tracking-widest text-amber-400">
              {roundLabel(minRound)}
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
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
}
