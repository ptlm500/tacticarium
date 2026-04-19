import { useState } from "react";
import { Stratagem } from "../../types/faction";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface Props {
  stratagems: Stratagem[];
  currentCP: number;
  usedThisPhase: string[];
  onUse: (stratagem: Stratagem, cpSpent: number) => void;
}

export function StratagemPanel({ stratagems, currentCP, usedThisPhase, onUse }: Props) {
  const [pending, setPending] = useState<Stratagem | null>(null);
  const [cpInput, setCpInput] = useState<number>(0);

  if (stratagems.length === 0) {
    return (
      <div className="mt-2 rounded-sm border border-border/40 bg-background/40 p-4 text-center text-sm text-muted-foreground">
        No stratagems available for this phase.
      </div>
    );
  }

  const openPrompt = (s: Stratagem) => {
    setPending(s);
    setCpInput(Math.min(s.cpCost, currentCP));
  };

  const closePrompt = () => setPending(null);

  const confirmUse = () => {
    if (!pending) return;
    const clamped = Math.max(0, Math.min(cpInput, currentCP));
    onUse(pending, clamped);
    setPending(null);
  };

  return (
    <>
      <div className="mt-2 max-h-80 space-y-2 overflow-y-auto pr-1">
        {stratagems.map((s) => (
          <div key={s.id} className="rounded-sm border border-border/60 bg-background/40 p-3">
            <div className="mb-1 flex items-start justify-between gap-2">
              <div>
                <h3 className="text-sm font-semibold text-foreground">{s.name}</h3>
                <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                  {s.type}
                </p>
              </div>
              <Badge
                variant={s.cpCost === 0 ? "secondary" : "default"}
                className={cn(
                  "font-mono uppercase tracking-widest",
                  s.cpCost === 0 && "border-emerald-500/40 bg-emerald-500/10 text-emerald-400",
                )}
              >
                {s.cpCost} CP
              </Badge>
            </div>
            {s.legend && <p className="mb-2 text-xs italic text-muted-foreground">{s.legend}</p>}
            <div className="mt-2 flex items-center justify-between gap-2">
              <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                {s.turn} | {s.phase}
              </span>
              <Button type="button" size="sm" onClick={() => openPrompt(s)}>
                Use
              </Button>
            </div>
          </div>
        ))}
      </div>

      <Dialog
        open={pending !== null}
        onOpenChange={(next) => {
          if (!next) closePrompt();
        }}
      >
        {pending && (
          <DialogContent className="sm:max-w-sm">
            <DialogHeader>
              <DialogTitle className="font-mono uppercase tracking-widest text-primary">
                {pending.name}
              </DialogTitle>
              <DialogDescription>
                Default cost: {pending.cpCost} CP. Adjust the amount to spend if a rule makes this
                stratagem more expensive, cheaper, or free.
              </DialogDescription>
            </DialogHeader>

            {usedThisPhase.includes(pending.id) && (
              <div
                role="alert"
                className="rounded-sm border border-amber-600/60 bg-amber-950/40 px-3 py-2 text-xs text-amber-200"
              >
                You've already used this stratagem this phase. Stratagems can normally only be used
                once per phase — only proceed if a rule allows an exception.
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="strat-cp-input" className="text-xs text-muted-foreground">
                CP to spend (you have {currentCP})
              </Label>
              <Input
                id="strat-cp-input"
                type="number"
                min={0}
                max={currentCP}
                value={cpInput}
                onChange={(e) => setCpInput(Math.max(0, Number(e.target.value)))}
                autoFocus
              />
              {cpInput > currentCP && (
                <p className="text-xs text-destructive">
                  You only have {currentCP} CP — cannot spend more than that.
                </p>
              )}
            </div>

            <DialogFooter className="gap-2 sm:gap-2">
              <Button type="button" variant="outline" onClick={closePrompt} className="flex-1">
                Cancel
              </Button>
              <Button
                type="button"
                onClick={confirmUse}
                disabled={cpInput > currentCP || cpInput < 0}
                className="flex-1"
              >
                Confirm
              </Button>
            </DialogFooter>
          </DialogContent>
        )}
      </Dialog>
    </>
  );
}
