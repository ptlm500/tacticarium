import { useQuery } from "@tanstack/react-query";
import { gamesApi } from "../../api/games";
import { queryKeys } from "../queryKeys";

export function useGameHistory(filters?: { myFaction?: string; opponentFaction?: string }) {
  return useQuery({
    queryKey: queryKeys.history.list(filters),
    queryFn: () => gamesApi.getHistory(filters),
  });
}

export function useUserStats() {
  return useQuery({
    queryKey: queryKeys.history.stats(),
    queryFn: () => gamesApi.getStats(),
  });
}
