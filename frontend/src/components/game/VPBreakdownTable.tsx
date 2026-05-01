import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
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
  const headClass = "text-center font-mono text-[10px] uppercase tracking-widest text-primary";

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-muted-foreground font-mono text-[10px] uppercase tracking-widest">
            Round
          </TableHead>
          <TableHead colSpan={2} className={cn(headClass, "border-l border-border/50")}>
            {myUsername}
          </TableHead>
          {opponentStats && (
            <TableHead colSpan={2} className={cn(headClass, "border-l border-border/50")}>
              {opponentUsername}
            </TableHead>
          )}
        </TableRow>
        <TableRow className="hover:bg-transparent">
          <TableHead />
          <TableHead className={cn(headClass, "border-l border-border/50")}>Pri</TableHead>
          <TableHead className={headClass}>Sec</TableHead>
          {opponentStats && (
            <>
              <TableHead className={cn(headClass, "border-l border-border/50")}>Pri</TableHead>
              <TableHead className={headClass}>Sec</TableHead>
            </>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rounds.map((r) => {
          const my = myStats.vpByRound[r] ?? { primary: 0, secondary: 0 };
          const opp = opponentStats?.vpByRound[r] ?? { primary: 0, secondary: 0 };
          return (
            <TableRow key={r}>
              <TableCell className="text-muted-foreground font-mono">R{r}</TableCell>
              <TableCell className="text-center border-l border-border/50">
                {my.primary || "-"}
              </TableCell>
              <TableCell className="text-center">{my.secondary || "-"}</TableCell>
              {opponentStats && (
                <>
                  <TableCell className="text-center border-l border-border/50">
                    {opp.primary || "-"}
                  </TableCell>
                  <TableCell className="text-center">{opp.secondary || "-"}</TableCell>
                </>
              )}
            </TableRow>
          );
        })}
        <TableRow>
          <TableCell className="text-muted-foreground font-mono">Paint</TableCell>
          <TableCell colSpan={2} className="text-center border-l border-border/50">
            {myStats.paint}
          </TableCell>
          {opponentStats && (
            <TableCell colSpan={2} className="text-center border-l border-border/50">
              {opponentStats.paint}
            </TableCell>
          )}
        </TableRow>
      </TableBody>
      <TableFooter className="bg-transparent">
        <TableRow className="hover:bg-transparent">
          <TableCell className="font-mono uppercase tracking-widest text-primary">Total</TableCell>
          <TableCell
            colSpan={2}
            className="text-center border-l border-border/50 text-lg font-bold text-primary"
          >
            {myStats.totalVP} VP
          </TableCell>
          {opponentStats && (
            <TableCell
              colSpan={2}
              className="text-center border-l border-border/50 text-lg font-bold"
            >
              {opponentStats.totalVP} VP
            </TableCell>
          )}
        </TableRow>
      </TableFooter>
    </Table>
  );
}
