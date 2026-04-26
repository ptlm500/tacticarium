import { QueryClient, QueryCache } from "@tanstack/react-query";
import { HttpError, clearToken } from "./api/client";

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        // Client-level retry handles transient 5xx and network errors with
        // exponential backoff; TanStack should not double-retry.
        retry: false,
        refetchOnWindowFocus: false,
        throwOnError: true,
      },
      mutations: {
        throwOnError: false,
      },
    },
    queryCache: new QueryCache({
      onError: (error) => {
        const isUnauthorized =
          (error instanceof HttpError && error.status === 401) || error.message.includes("401");
        if (isUnauthorized) {
          clearToken();
          window.location.href = "/login";
        }
      },
    }),
  });
}
