import { useState } from "react";
import type { PlayerState } from "../../types/game";
import { PlayerScoringHeatmap } from "./PlayerScoringHeatmap";
import { ScoringDetailModal, type CellSelection } from "./ScoringDetailModal";
import type { ScoringHeatmapData, VPCategory } from "./vpUtils";

interface Props {
  players: (PlayerState | null)[];
  data: ScoringHeatmapData;
  className?: string;
}

export function ScoringHeatmaps({ players, data, className }: Props) {
  const [selection, setSelection] = useState<CellSelection | null>(null);
  const present = players.filter((p): p is PlayerState => p !== null);

  if (data.rounds.length === 0 || present.length === 0) return null;

  return (
    <div className={className ?? "grid grid-cols-1 gap-3 sm:grid-cols-2"}>
      {present.map((p) => {
        const stats = data.statsByPlayerNumber[p.playerNumber];
        if (!stats) return null;
        return (
          <PlayerScoringHeatmap
            key={p.playerNumber}
            username={p.username}
            stats={stats}
            rounds={data.rounds}
            intensityMax={data.intensityMax}
            onCellClick={(round, category: VPCategory) =>
              setSelection({
                playerNumber: p.playerNumber,
                username: p.username,
                round,
                category,
              })
            }
          />
        );
      })}
      <ScoringDetailModal
        selection={selection}
        onClose={() => setSelection(null)}
        events={data.normalizedEvents}
      />
    </div>
  );
}
