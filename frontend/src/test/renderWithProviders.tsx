import { ReactElement } from "react";
import { render, RenderOptions } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "../hooks/useAuth";
import { User } from "../api/auth";
import { mockUser } from "./fixtures";

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        throwOnError: true,
      },
      mutations: {
        throwOnError: false,
      },
    },
  });
}

interface ProviderOptions {
  user?: User | null;
  route?: string;
  queryClient?: QueryClient;
}

export function renderWithProviders(
  ui: ReactElement,
  { user = mockUser, route = "/", queryClient }: ProviderOptions = {},
  renderOptions?: RenderOptions,
) {
  const testQueryClient = queryClient ?? createTestQueryClient();

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={testQueryClient}>
        <AuthContext.Provider
          value={{
            user: user ?? null,
            loading: false,
            login: () => {},
            logout: async () => {},
          }}
        >
          <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
        </AuthContext.Provider>
      </QueryClientProvider>
    );
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions });
}
