import type { PlayerState } from "../../types/game";
import { useStratagems } from "../../hooks/queries/useFactionQueries";
import { PlayerAvatar } from "./PlayerAvatar";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Badge } from "@/components/ui/badge";

interface Props {
  player: PlayerState;
  isActive: boolean;
}

export function SpectatorPlayerPanel({ player, isActive }: Props) {
  const totalVP = player.vpPrimary + player.vpSecondary + player.vpPaint;

  const secondaryHand = player.secondaryHand ?? [];
  const secondaryScored = player.secondaryScored ?? [];
  const secondaryDeck = player.secondaryDeck ?? [];
  const stratagemsUsed = player.stratagemsUsedThisPhase ?? [];

  const { data: stratagems } = useStratagems(player.factionId);
  const stratagemNameById = new Map((stratagems ?? []).map((s) => [s.id, s.name]));

  const label = `${player.username} — ${player.factionName || "Unknown faction"}`;

  return (
    <HUDFrame label={label}>
      <div className="space-y-4 py-1">
        <div className="flex items-center gap-3">
          <PlayerAvatar avatarUrl={player.avatarUrl} username={player.username} size="md" />
          <div className="font-mono text-sm text-foreground">{player.username}</div>
        </div>
        <div className="flex flex-wrap items-center gap-2 font-mono text-[10px] uppercase tracking-widest">
          {isActive && (
            <Badge
              variant="outline"
              className="border-primary/60 text-primary shadow-[0_0_10px_var(--primary)]"
            >
              Active Turn
            </Badge>
          )}
          {player.detachmentName && (
            <span className="text-muted-foreground">
              Detachment: <span className="text-foreground">{player.detachmentName}</span>
            </span>
          )}
        </div>

        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <Stat label="CP" value={player.cp} />
          <Stat label="Total VP" value={totalVP} highlight />
          <Stat label="Primary" value={player.vpPrimary} />
          <Stat label="Secondary" value={player.vpSecondary} />
          <Stat label="Paint" value={player.vpPaint} />
        </div>

        <div>
          <SectionHeading>
            Secondaries ({player.secondaryMode === "tactical" ? "Tactical" : "Fixed"})
            {player.secondaryMode === "tactical" && secondaryDeck.length > 0 && (
              <span className="ml-2 normal-case tracking-normal text-muted-foreground/70">
                — Deck: {secondaryDeck.length}
              </span>
            )}
          </SectionHeading>
          {secondaryHand.length === 0 ? (
            <p className="font-mono text-[11px] uppercase tracking-widest text-muted-foreground/70">
              No active secondaries
            </p>
          ) : (
            <ul className="space-y-2">
              {secondaryHand.map((s) => (
                <li key={s.id} className="rounded-sm border border-border/60 bg-background/40 p-2">
                  <div className="flex items-start justify-between gap-2">
                    <span className="text-sm font-medium text-foreground">{s.name}</span>
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">{s.text}</p>
                </li>
              ))}
            </ul>
          )}
        </div>

        {secondaryScored.length > 0 && (
          <div>
            <SectionHeading>Scored ({secondaryScored.length})</SectionHeading>
            <ul className="space-y-1 font-mono text-[11px] text-muted-foreground">
              {secondaryScored.map((s) => (
                <li key={s.id} className="flex justify-between gap-2">
                  <span className="truncate text-emerald-300/90">{s.name}</span>
                  {s.vpScored != null && s.vpScored > 0 && (
                    <span className="tabular-nums text-emerald-300">+{s.vpScored}</span>
                  )}
                </li>
              ))}
            </ul>
          </div>
        )}

        {stratagemsUsed.length > 0 && (
          <div>
            <SectionHeading>Stratagems This Phase</SectionHeading>
            <div className="flex flex-wrap gap-1">
              {stratagemsUsed.map((id) => (
                <Badge
                  key={id}
                  variant="outline"
                  className="font-mono text-[10px] uppercase tracking-widest"
                >
                  {stratagemNameById.get(id) ?? id}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </div>
    </HUDFrame>
  );
}

function Stat({ label, value, highlight }: { label: string; value: number; highlight?: boolean }) {
  return (
    <div className="rounded-sm border border-border/40 bg-background/40 px-2 py-1.5">
      <div className="font-mono text-[9px] uppercase tracking-widest text-muted-foreground">
        {label}
      </div>
      <div
        className={`font-mono text-lg tabular-nums ${
          highlight ? "text-primary" : "text-foreground"
        }`}
      >
        {value}
      </div>
    </div>
  );
}

function SectionHeading({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="mb-2 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
      {children}
    </h3>
  );
}
