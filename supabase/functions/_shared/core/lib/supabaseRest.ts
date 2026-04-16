export function getSupabaseConfig() {
  const baseUrl = Deno.env.get("SUPABASE_URL") ?? "http://127.0.0.1:54321";
  const apiKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY") ??
    Deno.env.get("SUPABASE_ANON_KEY") ?? "";
  return { baseUrl: baseUrl.replace(/\/$/, ""), apiKey };
}

export async function restRequest<T>(
  method: string,
  path: string,
  body?: unknown,
  extraHeaders: Record<string, string> = {},
): Promise<T> {
  const { baseUrl, apiKey } = getSupabaseConfig();
  if (!apiKey) {
    throw new Error(
      "Missing SUPABASE_SERVICE_ROLE_KEY or SUPABASE_ANON_KEY for Supabase REST calls.",
    );
  }

  const resp = await fetch(`${baseUrl}/rest/v1/${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      "apikey": apiKey,
      "Authorization": `Bearer ${apiKey}`,
      ...extraHeaders,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const text = await resp.text();
  const parsed = text ? JSON.parse(text) : null;
  if (!resp.ok) {
    throw new Error(`${method} ${path} failed: ${resp.status} ${text}`);
  }
  return parsed as T;
}
