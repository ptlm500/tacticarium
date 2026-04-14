import { useMutation, useQueryClient } from "@tanstack/react-query";
import { gamesApi } from "../../api/games";
import { GameSummary } from "../../types/game";
import { queryKeys } from "../queryKeys";

export function useCreateGame() {
  return useMutation({
    mutationFn: () => gamesApi.create(),
  });
}

export function useJoinGame() {
  return useMutation({
    mutationFn: (code: string) => gamesApi.join(code),
  });
}

export function useHideGame() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => gamesApi.hide(id),
    onSuccess: (_data, id) => {
      queryClient.setQueryData(queryKeys.games.list(), (old: GameSummary[] | undefined) =>
        old?.filter((g) => g.id !== id),
      );
      queryClient.setQueryData(queryKeys.history.list(), (old: GameSummary[] | undefined) =>
        old?.filter((g) => g.id !== id),
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.history.all });
    },
  });
}
