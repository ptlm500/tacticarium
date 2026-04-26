const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export function getToken(): string {
  return localStorage.getItem("token") || "";
}

export function clearToken() {
  localStorage.removeItem("token");
}

export class HttpError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "HttpError";
    this.status = status;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  if (!res.ok) {
    let message = `Request failed: ${res.status}`;
    const text = await res.text();
    if (text) {
      try {
        const body = JSON.parse(text);
        message = body.error || body.message || text;
      } catch {
        message = text;
      }
    }
    throw new HttpError(message, res.status);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

const MAX_GET_ATTEMPTS = 3;
const BASE_RETRY_DELAY_MS = 250;

async function getWithRetry<T>(path: string): Promise<T> {
  let lastError: unknown;
  for (let attempt = 0; attempt < MAX_GET_ATTEMPTS; attempt++) {
    try {
      return await request<T>(path);
    } catch (err) {
      lastError = err;
      const retryable = (err instanceof HttpError && err.status >= 500) || err instanceof TypeError;
      if (!retryable || attempt === MAX_GET_ATTEMPTS - 1) throw err;
      const delay = BASE_RETRY_DELAY_MS * Math.pow(3, attempt);
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }
  throw lastError;
}

export const api = {
  get: <T>(path: string) => getWithRetry<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    }),
};
