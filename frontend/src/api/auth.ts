import { api } from "./client";

export interface User {
  id: string;
  username: string;
  avatar?: string;
  createdAt: string;
}

export const authApi = {
  getMe: () => api.get<User>("/api/auth/me"),
  logout: () => api.post<void>("/api/auth/logout"),
  getLoginUrl: () => {
    const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:8080";
    return `${apiUrl}/api/auth/discord`;
  },
};
