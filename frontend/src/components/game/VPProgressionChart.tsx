import { LineChart, Line, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer } from "recharts";
import type { PlayerSummaryStats } from "./vpUtils";

interface Props {
  myStats: PlayerSummaryStats;
  opponentStats: PlayerSummaryStats | null;
  myUsername: string;
  opponentUsername: string | null;
  rounds: number[];
}

export function VPProgressionChart({
  myStats,
  opponentStats,
  myUsername,
  opponentUsername,
  rounds,
}: Props) {
  let myCumulative = 0;
  let oppCumulative = 0;

  const data = rounds.map((r) => {
    const my = myStats.vpByRound[r];
    if (my) myCumulative += my.primary + my.secondary;

    const opp = opponentStats?.vpByRound[r];
    if (opp) oppCumulative += opp.primary + opp.secondary;

    return {
      round: `R${r}`,
      [myUsername]: myCumulative,
      ...(opponentStats ? { [opponentUsername!]: oppCumulative } : {}),
    };
  });

  return (
    <div className="w-full h-56">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
          <XAxis dataKey="round" stroke="#9ca3af" fontSize={12} />
          <YAxis stroke="#9ca3af" fontSize={12} />
          <Tooltip
            contentStyle={{
              backgroundColor: "#1f2937",
              border: "1px solid #374151",
              borderRadius: "0.5rem",
              color: "#f3f4f6",
            }}
          />
          <Legend wrapperStyle={{ fontSize: 12, color: "#9ca3af" }} />
          <Line
            type="monotone"
            dataKey={myUsername}
            stroke="#818cf8"
            strokeWidth={2}
            dot={{ fill: "#818cf8", r: 4 }}
          />
          {opponentStats && (
            <Line
              type="monotone"
              dataKey={opponentUsername!}
              stroke="#f87171"
              strokeWidth={2}
              dot={{ fill: "#f87171", r: 4 }}
            />
          )}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
