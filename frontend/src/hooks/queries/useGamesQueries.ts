import { useQuery } from "@tanstack/react-query";
import { gamesApi } from "../../api/games";
import { queryKeys } from "../queryKeys";

export function useGameList() {
  return useQuery({
    queryKey: queryKeys.games.list(),
    queryFn: () => gamesApi.list(),
  });
}

export function useGame(id: string) {
  return useQuery({
    queryKey: queryKeys.games.detail(id),
    queryFn: () => gamesApi.get(id),
  });
}

export function useGameEvents(id: string) {
  return useQuery({
    queryKey: queryKeys.games.events(id),
    queryFn: () => gamesApi.getEvents(id),
  });
}
