import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AuthContext } from '../hooks/useAuth';
import { User } from '../api/auth';
import { mockUser } from './fixtures';

interface ProviderOptions {
  user?: User | null;
  route?: string;
}

export function renderWithProviders(
  ui: ReactElement,
  { user = mockUser, route = '/' }: ProviderOptions = {},
  renderOptions?: RenderOptions
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
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
    );
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions });
}
