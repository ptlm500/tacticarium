import { useState } from "react";
import type { Board, Objective } from "../../types/game";
import type { DeploymentPattern } from "../../types/mission";

interface Props {
  board: Board;
  pattern: DeploymentPattern | null;
  /** Called when an objective's control is cycled (player = 0 none / 1 / 2). */
  onSetControl: (objectiveIndex: number, player: number) => void;
}

// The 40kdc board is 60" wide × 44" deep. A small margin keeps strokes inside
// the SVG viewport.
const BOARD_W = 60;
const BOARD_H = 44;

interface Point {
  x: number;
  y: number;
}
interface RegionShape {
  type?: string;
  points?: Point[];
  width?: number;
  height?: number;
}
interface Region {
  player?: string;
  name?: string;
  color?: string;
  position?: Point;
  shape?: RegionShape;
}

/** Absolute polygon points for a region (position offset + shape). */
function regionPolygon(region: Region): Point[] {
  const pos = region.position ?? { x: 0, y: 0 };
  const shape = region.shape;
  if (!shape) return [];
  if (shape.type === "rectangle" && shape.width != null && shape.height != null) {
    return [
      { x: pos.x, y: pos.y },
      { x: pos.x + shape.width, y: pos.y },
      { x: pos.x + shape.width, y: pos.y + shape.height },
      { x: pos.x, y: pos.y + shape.height },
    ];
  }
  return (shape.points ?? []).map((p) => ({ x: pos.x + p.x, y: pos.y + p.y }));
}

function toPointsAttr(points: Point[]): string {
  return points.map((p) => `${p.x},${p.y}`).join(" ");
}

function asRegions(value: unknown): Region[] {
  return Array.isArray(value) ? (value as Region[]) : [];
}

function controlFill(controlledBy: number): string {
  if (controlledBy === 1) return "var(--primary)";
  if (controlledBy === 2) return "#f59e0b"; // amber — the opponent marker
  return "var(--muted-foreground)";
}

const ROLE_ABBR: Record<string, string> = {
  home: "H",
  central: "C",
  expansion: "E",
};

function controlLabel(controlledBy: number): string {
  if (controlledBy === 1) return "Player 1";
  if (controlledBy === 2) return "Player 2";
  return "no one";
}

export function BoardView({ board, pattern, onSetControl }: Props) {
  const [expanded, setExpanded] = useState(true);
  const objectives = board.objectives ?? [];
  const territories = asRegions(pattern?.territories);
  const zones = asRegions(pattern?.zones);

  const cycle = (obj: Objective) => {
    const next = (obj.controlledBy + 1) % 3;
    onSetControl(obj.index, next);
  };

  return (
    <section className="space-y-2">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between rounded-sm border border-border/60 bg-background/40 px-4 py-3 text-left transition-colors hover:border-primary/50"
      >
        <span className="font-mono text-sm uppercase tracking-widest text-primary">
          Battlefield
        </span>
        <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          Tap an objective to set control
        </span>
      </button>

      {expanded && objectives.length === 0 && (
        <p className="rounded-sm border border-border/60 bg-background/40 px-4 py-6 text-center font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          No deployment objectives placed
        </p>
      )}

      {expanded && objectives.length > 0 && (
        <div className="space-y-2 rounded-sm border border-border/60 bg-background/40 p-3">
          <svg
            viewBox={`-1 -1 ${BOARD_W + 2} ${BOARD_H + 2}`}
            className="w-full rounded-sm bg-background/60"
            role="img"
            aria-label="Battlefield objectives"
          >
            <rect
              x={0}
              y={0}
              width={BOARD_W}
              height={BOARD_H}
              fill="none"
              stroke="var(--border)"
              strokeWidth={0.3}
            />

            {territories.map((t, i) => (
              <polygon
                key={`terr-${i}`}
                points={toPointsAttr(regionPolygon(t))}
                fill={t.player === "attacker" ? "rgba(245,158,11,0.06)" : "rgba(59,130,246,0.06)"}
                stroke="none"
              />
            ))}

            {zones.map((z, i) => (
              <polygon
                key={`zone-${i}`}
                points={toPointsAttr(regionPolygon(z))}
                fill="none"
                stroke={z.color ?? "var(--muted-foreground)"}
                strokeWidth={0.3}
                strokeDasharray="1 1"
                opacity={0.6}
              />
            ))}

            {objectives.map((obj) => (
              <g
                key={obj.index}
                role="button"
                tabIndex={0}
                aria-label={`Objective ${obj.index + 1} (${obj.role}), controlled by ${controlLabel(
                  obj.controlledBy,
                )}`}
                className="cursor-pointer outline-none"
                onClick={() => cycle(obj)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    cycle(obj);
                  }
                }}
              >
                <circle
                  cx={obj.point.x}
                  cy={obj.point.y}
                  r={2.6}
                  fill={controlFill(obj.controlledBy)}
                  fillOpacity={obj.controlledBy === 0 ? 0.25 : 0.9}
                  stroke={controlFill(obj.controlledBy)}
                  strokeWidth={0.4}
                />
                <text
                  x={obj.point.x}
                  y={obj.point.y + 0.9}
                  textAnchor="middle"
                  fontSize={2.4}
                  fill="var(--background)"
                  className="select-none font-mono"
                >
                  {ROLE_ABBR[obj.role] ?? "?"}
                </text>
              </g>
            ))}
          </svg>

          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            <span className="flex items-center gap-1">
              <span className="inline-block size-2 rounded-full bg-primary" /> P1
            </span>
            <span className="flex items-center gap-1">
              <span
                className="inline-block size-2 rounded-full"
                style={{ background: "#f59e0b" }}
              />{" "}
              P2
            </span>
            <span>H = Home · C = Central · E = Expansion</span>
          </div>
        </div>
      )}
    </section>
  );
}
