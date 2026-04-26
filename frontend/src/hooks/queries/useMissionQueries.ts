import { useQuery } from "@tanstack/react-query";
import { missionsApi } from "../../api/missions";
import { queryKeys } from "../queryKeys";

// Missions are non-essential for the core game UI (the page can render
// without scoring rules / twist text) — degrade locally instead of bouncing
// the whole route to the QueryErrorBoundary.
export function useMissions(packId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.missions.list(packId!),
    queryFn: () => missionsApi.listMissions(packId!),
    enabled: !!packId,
    throwOnError: false,
  });
}

export function useMissionRules(packId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.missions.rules(packId!),
    queryFn: () => missionsApi.listRules(packId!),
    enabled: !!packId,
    throwOnError: false,
  });
}

export function useSecondaries(packId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.missions.secondaries(packId!),
    queryFn: () => missionsApi.listSecondaries(packId!),
    enabled: !!packId,
    throwOnError: false,
  });
}
