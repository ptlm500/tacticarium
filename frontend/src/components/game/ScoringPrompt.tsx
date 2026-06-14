import { ScoringAction } from "../../types/mission";
import { SecondaryCard } from "../../types/game";
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
  | { kind: "end_of_round_primary"; missionName: string; note: string }
  | {
      kind: "opponent_pending_secondary";
      secondaries: SecondaryCard[];
      opponentName: string;
    };

interface Props {
  items: ScoringPromptItem[];
  onScore: (
    category: string,
    delta: number,
    scoringSlot?: PrimaryScoringSlot,
    scoringRuleLabel?: string,
  ) => void;
  secondaryHand: SecondaryCard[];
  onConfirm: () => void;
  onCancel: () => void;
  title?: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
}

export function ScoringPrompt({
  items,
  onScore,
  secondaryHand,
  onConfirm,
  onCancel,
  title = "Scoring Reminder",
  description = "Before advancing, check if you need to score.",
  confirmLabel = "I've scored, continue",
  cancelLabel = "Let me score first",
}: Props) {
  return (
    <ReminderPrompt
      title={title}
      description={description}
      confirmLabel={confirmLabel}
      cancelLabel={cancelLabel}
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
          {item.kind === "secondary" && <SecondaryReminder secondaryHand={secondaryHand} />}
          {item.kind === "opponent_pending_secondary" && (
            <OpponentPendingSecondaryReminder
              secondaries={item.secondaries}
              opponentName={item.opponentName}
            />
          )}
        </div>
      ))}
    </ReminderPrompt>
  );
}

function OpponentPendingSecondaryReminder({
  secondaries,
  opponentName,
}: {
  secondaries: SecondaryCard[];
  opponentName: string;
}) {
  return (
    <div
      className="rounded-sm border border-amber-500/40 bg-amber-500/10 p-3"
      data-testid="opponent-pending-secondary"
    >
      <h3 className="font-mono text-sm uppercase tracking-widest text-amber-400">
        Wait for {opponentName} to score
      </h3>
      <p className="mt-1 text-xs text-foreground/80">
        Your opponent has secondaries that may score at the end of <em>your</em> turn. Confirm they
        have scored before continuing.
      </p>
      <ul className="mt-2 space-y-1 text-xs">
        {secondaries.map((s) => (
          <li key={s.id} className="font-medium text-foreground">
            • {s.name}
          </li>
        ))}
      </ul>
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

function SecondaryReminder({ secondaryHand }: { secondaryHand: SecondaryCard[] }) {
  return (
    <div className="rounded-sm border border-emerald-500/40 bg-emerald-500/10 p-3">
      <h3 className="font-mono text-sm uppercase tracking-widest text-emerald-400">
        Score Secondaries
      </h3>
      {secondaryHand.length === 0 ? (
        <p className="mt-1 text-xs text-foreground/80">No active secondary missions.</p>
      ) : (
        <div className="mt-2 space-y-3">
          {secondaryHand.map((s) => (
            <div key={s.id}>
              <span className="text-xs font-medium text-foreground">{s.name}</span>
              <p className="mt-1 text-xs text-muted-foreground">{s.text}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
