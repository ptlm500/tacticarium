import type { PlayerSummaryStats } from "./vpUtils";

interface Props {
  myStats: PlayerSummaryStats;
  opponentStats: PlayerSummaryStats | null;
  myUsername: string;
  opponentUsername: string | null;
  rounds: number[];
}

export function VPBreakdownTable({
  myStats,
  opponentStats,
  myUsername,
  opponentUsername,
  rounds,
}: Props) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-700">
            <th className="text-left py-2 text-gray-400 font-medium">Round</th>
            <th className="text-center py-2 text-gray-400 font-medium" colSpan={3}>
              {myUsername}
            </th>
            {opponentStats && (
              <th className="text-center py-2 text-gray-400 font-medium" colSpan={3}>
                {opponentUsername}
              </th>
            )}
          </tr>
          <tr className="border-b border-gray-600 text-xs text-gray-500">
            <th></th>
            <th className="py-1">Pri</th>
            <th className="py-1">Sec</th>
            <th className="py-1">Gam</th>
            {opponentStats && (
              <>
                <th className="py-1">Pri</th>
                <th className="py-1">Sec</th>
                <th className="py-1">Gam</th>
              </>
            )}
          </tr>
        </thead>
        <tbody>
          {rounds.map((r) => {
            const my = myStats.vpByRound[r] ?? { primary: 0, secondary: 0, gambit: 0 };
            const opp = opponentStats?.vpByRound[r] ?? { primary: 0, secondary: 0, gambit: 0 };
            return (
              <tr key={r} className="border-b border-gray-700/50">
                <td className="py-2 text-gray-400">R{r}</td>
                <td className="py-2 text-center">{my.primary || "-"}</td>
                <td className="py-2 text-center">{my.secondary || "-"}</td>
                <td className="py-2 text-center">{my.gambit || "-"}</td>
                {opponentStats && (
                  <>
                    <td className="py-2 text-center">{opp.primary || "-"}</td>
                    <td className="py-2 text-center">{opp.secondary || "-"}</td>
                    <td className="py-2 text-center">{opp.gambit || "-"}</td>
                  </>
                )}
              </tr>
            );
          })}
          {/* Paint row */}
          <tr className="border-b border-gray-700/50">
            <td className="py-2 text-gray-400">Paint</td>
            <td className="py-2 text-center" colSpan={3}>
              {myStats.paint}
            </td>
            {opponentStats && (
              <td className="py-2 text-center" colSpan={3}>
                {opponentStats.paint}
              </td>
            )}
          </tr>
          {/* Total row */}
          <tr className="font-bold text-base">
            <td className="py-2 text-gray-300">Total</td>
            <td className="py-2 text-center" colSpan={3}>
              {myStats.totalVP} VP
            </td>
            {opponentStats && (
              <td className="py-2 text-center" colSpan={3}>
                {opponentStats.totalVP} VP
              </td>
            )}
          </tr>
        </tbody>
      </table>
    </div>
  );
}
