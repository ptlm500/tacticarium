import { api } from "./client";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export interface AdminUser {
  githubId: string;
  githubUser: string;
}

export const authApi = {
  getMe: () => api.get<AdminUser>("/api/admin/me"),
  getLoginUrl: () => `${API_URL}/api/auth/github`,
};
