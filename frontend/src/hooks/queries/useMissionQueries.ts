import { useQuery } from "@tanstack/react-query";
import { missionsApi } from "../../api/missions";
import { queryKeys } from "../queryKeys";

// Reference data is non-essential for the core game UI (the page can render
// without it) — degrade locally instead of bouncing the whole route to the
// QueryErrorBoundary.

export function useMissions() {
  return useQuery({
    queryKey: queryKeys.missions.list(),
    queryFn: () => missionsApi.listMissions(),
    throwOnError: false,
  });
}

export function useForceDispositions() {
  return useQuery({
    queryKey: queryKeys.missions.forceDispositions(),
    queryFn: () => missionsApi.listForceDispositions(),
    throwOnError: false,
  });
}

export function useMissionMatchups() {
  return useQuery({
    queryKey: queryKeys.missions.matchups(),
    queryFn: () => missionsApi.listMissionMatchups(),
    throwOnError: false,
  });
}

export function useSecondaryCards() {
  return useQuery({
    queryKey: queryKeys.missions.secondaryCards(),
    queryFn: () => missionsApi.listSecondaryCards(),
    throwOnError: false,
  });
}

export function useDeploymentPatterns() {
  return useQuery({
    queryKey: queryKeys.missions.deploymentPatterns(),
    queryFn: () => missionsApi.listDeploymentPatterns(),
    throwOnError: false,
  });
}
