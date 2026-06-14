import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Mission } from "../../types/mission";
import { Badge } from "@/components/ui/badge";

interface Props {
  mission: Mission | null;
}

export function MissionInfo({ mission }: Props) {
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
                <div className="flex flex-wrap gap-2">
                  <Badge
                    variant="outline"
                    className="border-primary/40 bg-primary/10 font-mono uppercase tracking-widest text-primary"
                  >
                    {mission.vpPerRoundCap} VP / Round
                  </Badge>
                  <Badge
                    variant="outline"
                    className="border-primary/40 bg-primary/10 font-mono uppercase tracking-widest text-primary"
                  >
                    {mission.vpPerGameCap} VP / Game
                  </Badge>
                </div>
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
