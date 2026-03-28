import { useState, useEffect, useCallback, createContext, useContext } from 'react';
import { authApi, User } from '../api/auth';
import { getToken, clearToken } from '../api/client';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: () => void;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
  login: () => {},
  logout: async () => {},
});

export function useAuthProvider() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!getToken()) {
      setLoading(false);
      return;
    }
    authApi
      .getMe()
      .then(setUser)
      .catch(() => {
        clearToken();
        setUser(null);
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(() => {
    window.location.href = authApi.getLoginUrl();
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout().catch(() => {});
    clearToken();
    setUser(null);
  }, []);

  return { user, loading, login, logout };
}

export function useAuth() {
  return useContext(AuthContext);
}
