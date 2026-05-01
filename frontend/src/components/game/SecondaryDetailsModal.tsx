import { ActiveSecondary } from "../../types/game";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface Props {
  secondary: ActiveSecondary | null;
  onClose: () => void;
}

export function SecondaryDetailsModal({ secondary, onClose }: Props) {
  return (
    <Dialog
      open={secondary !== null}
      onOpenChange={(next) => {
        if (!next) onClose();
      }}
    >
      {secondary && (
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="font-mono uppercase tracking-widest text-primary">
              {secondary.name}
            </DialogTitle>
            <DialogDescription className="font-mono text-[10px] uppercase tracking-widest">
              {secondary.maxVp} VP max{secondary.isFixed ? " · Fixed" : ""}
            </DialogDescription>
          </DialogHeader>

          <p className="whitespace-pre-line text-sm text-foreground/90">{secondary.description}</p>

          {secondary.drawRestriction && (
            <div className="rounded-sm border border-amber-500/40 bg-amber-500/10 p-2 text-xs text-amber-200">
              <span className="font-mono uppercase tracking-widest text-amber-300">
                When Drawn (Round {secondary.drawRestriction.round})
              </span>{" "}
              —{" "}
              {secondary.drawRestriction.mode === "mandatory"
                ? "automatically shuffled back into the deck"
                : "may be shuffled back into the deck"}
            </div>
          )}

          {secondary.scoringOptions && secondary.scoringOptions.length > 0 && (
            <div className="space-y-1">
              <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Scoring:
              </p>
              <ul className="space-y-1">
                {secondary.scoringOptions.map((opt, i) => (
                  <li key={i} className="flex items-center gap-2 text-xs">
                    <Badge
                      variant="outline"
                      className="border-primary/40 bg-primary/10 font-mono uppercase tracking-widest text-primary"
                    >
                      +{opt.vp} VP
                    </Badge>
                    <span className="text-foreground/80">{opt.label}</span>
                    {opt.mode && (
                      <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                        ({opt.mode})
                      </span>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </DialogContent>
      )}
    </Dialog>
  );
}
