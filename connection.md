# Connect the AI agent (Vapi + Supabase)

This guide wires your **Vapi assistant** to this repo’s **Supabase Edge Function** webhook so tool and function calls hit `dispatch_dental_action` and the dental router in `supabase/functions/webhook`.

## What gets connected

| Piece | Role |
|--------|------|
| **Vapi assistant** | Voice agent; `serverUrl` must point at your webhook. |
| **Function tool** | e.g. `dispatch_dental_action` — Vapi POSTs tool payloads to `serverUrl`. |
| **Supabase Edge Function `webhook`** | Handles Vapi messages at `/functions/v1/webhook`. |

**Hosted URL shape (production):**

```text
https://<SUPABASE_PROJECT_REF>.supabase.co/functions/v1/webhook
```

**Local dev** uses `http://127.0.0.1:54321/functions/v1/webhook` unless you tunnel (Tailscale Funnel, etc.); then Vapi’s `serverUrl` should be your **public HTTPS** base + `/functions/v1/webhook`.

---

## Prerequisites

1. **Supabase CLI** — `supabase` on `PATH` (or `npx supabase`).
2. **Vapi CLI** — [install](https://docs.vapi.ai/cli), then `vapi login` (writes `~/.vapi-cli.yaml` with `api_key`).
3. A **Vapi assistant** you own — get its id: `vapi assistant list`.
4. **Linked Supabase project** — `supabase link --project-ref <ref>` *or* set `SUPABASE_PROJECT_REF` in `.env` (see `.env.example`).

---

## 1. Environment variables

Copy `.env.example` to `.env` and set at least:

| Variable | Required for | Notes |
|----------|----------------|------|
| `VAPI_ASSISTANT_ID` | Connect scripts | UUID from `vapi assistant list`. |
| `VAPI_API_KEY` | Patching assistant / tools | Optional if `vapi login` already set the key in `~/.vapi-cli.yaml`. |
| `SUPABASE_PROJECT_REF` | Deploy + hosted `serverUrl` | 20-char ref; or rely on `supabase/.temp/project-ref` after `supabase link`. |

Optional:

| Variable | Purpose |
|----------|---------|
| `CONNECT_ASSISTANT_MODE` | `full` (default): push `vapi/riley-assistant.json`. `server`: only set `serverUrl` on the assistant. |
| `PUBLIC_BASE_URL` | Tunnel origin **without** trailing slash; Vapi `serverUrl` = `${PUBLIC_BASE_URL}/functions/v1/webhook`. |
| `SKIP_TOOL_SYNC` | Default `1` in connect flow — skips Bearer header sync on tools when webhook has no auth. |
| `TOOL_API_KEY` | Only if you enable auth on the webhook and use `scripts/vapi/sync-tool-auth.sh`. |

---

## 2. One-command connect (hosted Supabase)

From the **repository root**:

```bash
./scripts/vapi/connect-supabase-vapi.sh
```

This script:

1. Deploys the `webhook` Edge Function to your linked project (`--no-verify-jwt`).
2. Sets `serverUrl` in `vapi/riley-assistant.json` to  
   `https://<SUPABASE_PROJECT_REF>.supabase.co/functions/v1/webhook`.
3. **If `CONNECT_ASSISTANT_MODE=full` (default):** runs `scripts/vapi/push-assistant-and-sync-tools.sh` to PATCH the assistant from `riley-assistant.json` (with `SKIP_TOOL_SYNC` defaulting to skip tool header sync).
4. **If `CONNECT_ASSISTANT_MODE=server`:** only PATCHes `serverUrl` on the assistant — leaves prompt/model/tools as-is.

Example — keep your existing dashboard agent, only point webhook:

```bash
export CONNECT_ASSISTANT_MODE=server
./scripts/vapi/connect-supabase-vapi.sh
```

---

## 3. Link an existing assistant (alternate script)

If you already ran the deploy step and only need to merge config or adjust the dispatch tool:

```bash
export VAPI_ASSISTANT_ID='<uuid>'
# optional: export VAPI_DISPATCH_TOOL_ID='...'
./scripts/vapi/configure-riley-supabase.sh
```

`scripts/vapi/link-existing-assistant.sh` is a thin wrapper that calls `configure-riley-supabase.sh`.

---

## 4. Local development

1. Start stack and serve the function:

   ```bash
   supabase start
   ./scripts/run-vapi-local.sh
   ```

2. Point Vapi at a URL Vapi’s cloud can reach:
   - Use **Tailscale Funnel** (or similar) and set `PUBLIC_BASE_URL` to the HTTPS origin.
   - See `./scripts/tailscale-expose-vapi.sh` for reference commands.

3. In the Vapi dashboard (or via API), set the assistant **`serverUrl`** to:

   ```text
   <PUBLIC_BASE_URL>/functions/v1/webhook
   ```

4. Health check:

   ```bash
   ./scripts/vapi-healthcheck.sh --public
   # or
   ./scripts/vapi-healthcheck.sh --url https://<host>
   ```

---

## 5. Credentials and security (read this)

- While developing, the Edge webhook may run **without inbound JWT verification** (`--no-verify-jwt`) so Vapi can POST without Supabase JWTs. **Do not treat the public URL as production-safe** until you add proper auth and lock down deployment flags.
- Vapi **“provider: supabase”** credentials in their UI are aimed at **Supabase Storage (S3-style)**, not at calling Edge Functions. See `vapi/CREDENTIALS.txt` and [Vapi Supabase provider docs](https://docs.vapi.ai/providers/cloud/supabase).
- If you later require `Authorization` on the webhook, set `TOOL_API_KEY` and run `scripts/vapi/sync-tool-auth.sh` so function tools send the Bearer header.

---

## 6. Quick verification checklist

- [ ] `vapi assistant get <VAPI_ASSISTANT_ID>` shows `serverUrl` matching your webhook.
- [ ] `GET` on the webhook URL returns JSON with `ok: true` (healthcheck script does this).
- [ ] A test call routes tool traffic; watch Supabase function logs if something fails.

---

## 7. Troubleshooting

| Symptom | What to check |
|---------|----------------|
| `Missing SUPABASE_PROJECT_REF` | `supabase link` or set `SUPABASE_PROJECT_REF` in `.env`. |
| `Missing VAPI_ASSISTANT_ID` | `export VAPI_ASSISTANT_ID=...` from `vapi assistant list`. |
| HTTP 401/403 on webhook | JWT or `TOOL_API_KEY` mismatch — align with `sync-tool-auth` or disable strict auth for dev. |
| Tools never fire | `serverUrl` wrong, tunnel down, or assistant not using the function tool id you expect (`VAPI_DISPATCH_TOOL_ID` in configure script). |

For more context on files and scripts, see `README.md` and `vapi/README.md`.
