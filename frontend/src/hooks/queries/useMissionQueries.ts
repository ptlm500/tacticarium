import { useQuery } from "@tanstack/react-query";
import { missionsApi } from "../../api/missions";
import { queryKeys } from "../queryKeys";

export function useMissions(packId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.missions.list(packId!),
    queryFn: () => missionsApi.listMissions(packId!),
    enabled: !!packId,
  });
}

export function useMissionRules(packId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.missions.rules(packId!),
    queryFn: () => missionsApi.listRules(packId!),
    enabled: !!packId,
  });
}

export function useSecondaries(packId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.missions.secondaries(packId!),
    queryFn: () => missionsApi.listSecondaries(packId!),
    enabled: !!packId,
  });
}
