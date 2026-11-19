import "server-only";

export async function apiFetch(path: string, cookie?: string): Promise<Response> {
  const base = process.env.API_INTERNAL_URL ?? "http://localhost:8080";
  return fetch(`${base}${path}`, {
    headers: cookie ? { cookie } : undefined,
    cache: "no-store",
  });
}

export async function apiJSON<T>(path: string, cookie?: string): Promise<T> {
  const response = await apiFetch(path, cookie);
  if (!response.ok) {
    throw new Error(`API ${path} returned ${response.status}`);
  }
  return (await response.json()) as T;
}
