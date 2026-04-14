import { QueryClient, QueryCache } from "@tanstack/react-query";
import { clearToken } from "./api/client";

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
        throwOnError: true,
      },
      mutations: {
        throwOnError: false,
      },
    },
    queryCache: new QueryCache({
      onError: (error) => {
        if (error.message.includes("401")) {
          clearToken();
          window.location.href = "/login";
        }
      },
    }),
  });
}
