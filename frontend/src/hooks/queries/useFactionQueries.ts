import { useQuery } from "@tanstack/react-query";
import { factionsApi } from "../../api/factions";
import { queryKeys } from "../queryKeys";

export function useFactions() {
  return useQuery({
    queryKey: queryKeys.factions.list(),
    queryFn: () => factionsApi.list(),
  });
}

export function useDetachments(factionId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.factions.detachments(factionId!),
    queryFn: () => factionsApi.getDetachments(factionId!),
    enabled: !!factionId,
  });
}

export function useStratagems(factionId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.factions.stratagems(factionId!),
    queryFn: () => factionsApi.getStratagems(factionId!),
    enabled: !!factionId,
    // Stratagems are non-essential for the core game UI — degrade locally
    // instead of bouncing the whole route to the QueryErrorBoundary.
    throwOnError: false,
  });
}
