// frontend/src/api.ts
export const API_BASE =
  (import.meta as any).env?.VITE_API_BASE?.replace(/\/$/, '') ?? 'http://localhost:8080';

type LoginResponse = {
  access_token: string;
  user: { id: number; email: string };
};

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${API_BASE}/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error(`Login failed (${res.status})`);
  return res.json();
}

export async function listApplications(token: string) {
  const res = await fetch(`${API_BASE}/v1/applications`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error(`Apps failed (${res.status})`);
  return res.json() as Promise<{ items: Array<{ id: number; title: string; company: string; location: string }> }>;
}
