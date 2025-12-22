import { API_BASE_URL } from "./env";
import { clearAccessToken, getAccessToken } from "./storage";

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: "GET" });
}

export async function apiPost<T>(path: string, body?: unknown, idempotencyKey?: string): Promise<T> {
  return apiRequest<T>(path, {
    method: "POST",
    body: body ? JSON.stringify(body) : undefined,
    headers: {
      "Content-Type": "application/json",
      ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
    },
  });
}

async function apiRequest<T>(path: string, init: RequestInit): Promise<T> {
  const token = getAccessToken();
  const url = buildUrl(path);

  const res = await fetch(url, {
    ...init,
    headers: {
      ...(init.headers || {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  const text = await res.text();
  const data = text ? JSON.parse(text) : null;


  if (!res.ok) {
    if (res.status === 401) {
      // session revoked (e.g., Toss unlink callback). Force re-login.
      clearAccessToken();
      if (typeof window !== "undefined") {
        const path = window.location.pathname || "";
        if (!path.startsWith("/login")) {
          window.location.href = "/login?reason=unlinked";
        }
      }
    }
    const msg = data?.error || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return data as T;
}

function buildUrl(path: string) {
  if (API_BASE_URL) return `${API_BASE_URL}${path}`;
  return path; // same-origin
}
