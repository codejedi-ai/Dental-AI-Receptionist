# AI Reception (Vapi + Go)

Everything runs **on your computer**: databases in Docker, **Go API only on localhost**. Your **tailnet / Tailscale Funnel** does not run the app — it only **publishes** `127.0.0.1` to the internet so **Vapi’s cloud** can `POST` tool webhooks to you.

## Layout

```
├── db/                 # PostgreSQL + MongoDB init scripts
├── vapi/               # Assistant JSON, deploy-vapi.sh, sync-tool-auth.sh
├── vapi-backend/       # Go — /api/tools (local HTTP only)
└── engineering-notebook/
```

## End-to-end: local Go → Funnel → Vapi CLI

Do this **on the same machine** that runs Go (a node in your tailnet with Funnel allowed in admin policy).

### 1) Databases (Docker — only these use containers)

```bash
docker compose up -d postgres mongo
# Optional: ./launch-vapi.sh   # DBs + Excel seed via local Python
```

### 2) Go backend (always local)

```bash
./run-vapi-local.sh
```

Listens on **`127.0.0.1:8080`** by default (not reachable from other PCs without a tunnel).

### 3) Expose localhost to the internet with Tailscale Funnel

```bash
./tailscale-expose-vapi.sh   # prints exact commands; default port 8080
```

Typical public setup:

```bash
sudo tailscale funnel --bg --yes 8080
```

Your tool base URL will look like **`https://<your-machine>.<tailnet>.ts.net`** (check `tailscale funnel status`).

### 4) Point Vapi at your tools URL

Set the assistant **`serverUrl`** to:

```text
https://<your-machine>.<tailnet>.ts.net/api/tools
```

Edit `vapi/riley-assistant.json` (`serverUrl` field), then deploy with the **Vapi CLI**:

```bash
curl -sSL https://vapi.ai/install.sh | bash
vapi login
export VAPI_API_KEY="..."   # from Vapi dashboard

./vapi/deploy-vapi.sh
```

Or patch only the assistant:

```bash
vapi assistant update 450435e9-4562-4ddd-8429-54584d3285a7 --config vapi/riley-assistant.json
```

(Replace with your assistant id if different.)

### 5) Tool authentication (Bearer)

If you set **`TOOL_API_KEY`** in `vapi-backend/.env` (or repo `.env`), push the same secret to Vapi’s tool server headers:

```bash
./vapi/sync-tool-auth.sh
```

### 6) Verify

```bash
./vapi-healthcheck.sh --public   # needs PUBLIC_BASE_URL=https://<same-as-serverUrl-without-/api/tools>
```

---

## Other commands

| Script | Purpose |
|--------|---------|
| `./build-test-local.sh` | `go build` + `go test` for `vapi-backend` |
| `./stop-vapi-go.sh` | Stop local Go processes |
| `./setup-local-databases.sh` | Init DBs without Docker (optional) |

## Stack

- **Go** — local tool server only  
- **PostgreSQL + MongoDB** — Docker Compose  
- **Tailscale Funnel** — on the **same host** as Go; exposes `127.0.0.1:8080` as HTTPS  
- **Vapi CLI** — deploy assistant + `serverUrl`; **sync-tool-auth** — Bearer for tools  

## License

MIT
