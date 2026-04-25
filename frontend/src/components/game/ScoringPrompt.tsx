import { ScoringAction } from "../../types/mission";
import { ActiveSecondary } from "../../types/game";
import { PrimaryScoringSlot } from "../../types/scoring";
import { ReminderPrompt } from "./ReminderPrompt";
import { Button } from "@/components/ui/button";

export type ScoringPromptItem =
  | {
      kind: "primary";
      missionName: string;
      scoringRules: ScoringAction[];
      currentRound: number;
      scoringSlot: PrimaryScoringSlot;
    }
  | { kind: "secondary" }
  | { kind: "fixed_secondary"; secondaries: ActiveSecondary[] }
  | { kind: "end_of_round_primary"; missionName: string; note: string };

interface Props {
  items: ScoringPromptItem[];
  onScore: (
    category: string,
    delta: number,
    scoringSlot?: PrimaryScoringSlot,
    scoringRuleLabel?: string,
  ) => void;
  activeSecondaries: ActiveSecondary[];
  onAchieveSecondary: (id: string, vp: number) => void;
  onDiscardSecondary: (id: string, free: boolean) => void;
  canGainCP: boolean;
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
  canGainCP,
  onScoreFixedVP,
  onConfirm,
  onCancel,
}: Props) {
  return (
    <ReminderPrompt
      title="Scoring Reminder"
      description="Before advancing, check if you need to score."
      confirmLabel="I've scored, continue"
      cancelLabel="Let me score first"
      onConfirm={onConfirm}
      onCancel={onCancel}
    >
      {items.map((item, i) => (
        <div key={i}>
          {item.kind === "primary" && (
            <PrimaryReminder
              missionName={item.missionName}
              scoringRules={item.scoringRules}
              currentRound={item.currentRound}
              onScore={(vp, label) => onScore("primary", vp, item.scoringSlot, label)}
            />
          )}
          {item.kind === "end_of_round_primary" && (
            <div className="rounded-sm border border-primary/40 bg-primary/10 p-3">
              <h3 className="font-mono text-sm uppercase tracking-widest text-primary">
                Primary Mission — {item.missionName}
              </h3>
              <p className="mt-1 text-xs text-foreground/80">{item.note}</p>
            </div>
          )}
          {item.kind === "fixed_secondary" && (
            <FixedSecondaryReminder secondaries={item.secondaries} onScore={onScoreFixedVP} />
          )}
          {item.kind === "secondary" && (
            <SecondaryReminder
              activeSecondaries={activeSecondaries}
              onAchieve={onAchieveSecondary}
              onDiscard={onDiscardSecondary}
              canGainCP={canGainCP}
            />
          )}
        </div>
      ))}
    </ReminderPrompt>
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
  onScore: (vp: number, ruleLabel: string) => void;
}) {
  return (
    <div className="rounded-sm border border-primary/40 bg-primary/10 p-3">
      <h3 className="font-mono text-sm uppercase tracking-widest text-primary">
        Score Primary — {missionName}
      </h3>
      <div className="mt-2 flex flex-wrap gap-2">
        {scoringRules.map((action, i) => {
          const locked = action.minRound != null && currentRound < action.minRound;
          return (
            <Button
              key={i}
              type="button"
              size="sm"
              data-testid="scoring-prompt-primary-btn"
              onClick={() => onScore(action.vp, action.label)}
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
    <div className="rounded-sm border border-emerald-500/40 bg-emerald-500/10 p-3">
      <h3 className="font-mono text-sm uppercase tracking-widest text-emerald-400">
        Score / Discard Secondaries
      </h3>
      {activeSecondaries.length === 0 ? (
        <p className="mt-1 text-xs text-foreground/80">No active secondary missions.</p>
      ) : (
        <div className="mt-2 space-y-3">
          {activeSecondaries.map((s) => {
            const opts = (s.scoringOptions ?? []).filter((o) => !o.mode || o.mode === "tactical");
            return (
              <div key={s.id}>
                <span className="text-xs font-medium text-foreground">{s.name}</span>
                <div className="mt-1 flex flex-wrap gap-1">
                  {opts.map((opt, i) => (
                    <Button
                      key={i}
                      type="button"
                      size="sm"
                      onClick={() => onAchieve(s.id, opt.vp)}
                      title={opt.label}
                      className="bg-emerald-600 text-white hover:bg-emerald-700"
                    >
                      {opt.label} +{opt.vp}
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
                  {canGainCP && (
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => onDiscard(s.id, false)}
                      className="bg-teal-700 text-white hover:bg-teal-800"
                    >
                      +1CP
                    </Button>
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

function FixedSecondaryReminder({
  secondaries,
  onScore,
}: {
  secondaries: ActiveSecondary[];
  onScore: (vp: number) => void;
}) {
  return (
    <div className="rounded-sm border border-emerald-500/40 bg-emerald-500/10 p-3">
      <h3 className="font-mono text-sm uppercase tracking-widest text-emerald-400">
        Score Fixed Secondaries
      </h3>
      <div className="mt-2 space-y-3">
        {secondaries.map((s) => {
          const opts = (s.scoringOptions ?? []).filter((o) => !o.mode || o.mode === "fixed");
          return (
            <div key={s.id}>
              <p className="text-xs font-medium text-foreground">{s.name}</p>
              <div className="mt-1 flex flex-wrap gap-1">
                {opts.map((opt, i) => (
                  <Button
                    key={i}
                    type="button"
                    size="sm"
                    onClick={() => onScore(opt.vp)}
                    title={opt.label}
                    className="bg-emerald-600 text-white hover:bg-emerald-700"
                  >
                    {opt.label} +{opt.vp}VP
                  </Button>
                ))}
              </div>
              <p className="mt-1 text-xs text-muted-foreground">max {s.maxVp} VP total</p>
            </div>
          );
        })}
      </div>
    </div>
  );
}
