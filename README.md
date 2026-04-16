# AI Reception (Vapi + Supabase Serverless)

This project now runs with a serverless runtime for tool/webhook handling via Supabase Edge Functions.

## Runtime Architecture

- Supabase Edge Function: `supabase/functions/webhook`
- Vapi tool server URL: `https://<your-host>/functions/v1/webhook`
- PostgreSQL data store: `appointments`, `patients`, `dentists`

There is no Go runtime required for the webhook/tool execution path.

## Quick Start

1. Start local services:

```bash
supabase start
```

2. Run local webhook function:

```bash
./scripts/run-vapi-local.sh
```

3. Set Vapi assistant `serverUrl` to:

```text
https://<your-machine>.<tailnet>.ts.net/functions/v1/webhook
```

4. Push assistant and sync tool auth:

```bash
./scripts/vapi/push-assistant-and-sync-tools.sh
```

Or do Supabase + Vapi CLI wiring in one step:

```bash
./scripts/vapi/connect-supabase-vapi.sh
```

5. Verify health and tool-call reachability:

```bash
./scripts/vapi-healthcheck.sh --public
```

## Useful Commands

- `./scripts/launch-vapi.sh`: Bring up local data services and seed schedule
- `./scripts/tailscale-expose-vapi.sh`: Print Funnel/Serve commands
- `./scripts/build-test-local.sh`: Validate serverless local runtime prerequisites

## Notes

- If `TOOL_API_KEY` is set in `.env`, webhook requests must include `Authorization: Bearer <TOOL_API_KEY>`.
- `PUBLIC_BASE_URL` should be the tunnel base URL without trailing slash.

## License

MIT
