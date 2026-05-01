import { ScoringOption } from "../../types/game";

export type Pile = "deck" | "active" | "achieved" | "discarded";

export const ACTIVE_PILE_LIMIT = 2;

export function filterOptions(
  options: ScoringOption[] | null | undefined,
  mode: string,
): ScoringOption[] {
  if (!options || options.length === 0) return [];
  return options.filter((o) => !o.mode || o.mode === mode);
}
