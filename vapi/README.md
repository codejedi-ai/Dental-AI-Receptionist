# Vapi Configuration

This folder contains the Vapi assistant configuration for Riley, the Smile Dental Clinic AI receptionist, (JSON and docs). Shell scripts live under **`../scripts/vapi/`**.

## Files

| File | Purpose |
|------|---------|
| `riley-assistant.json` | **Deploy source** — Full assistant config (prompt, tools, voice, transcriber, serverUrl) |
| `riley-current.json` | Live snapshot from Vapi (for reference/diff) |
| `analysis-plan.json` | Call analysis evaluation rubric (Checklist-based post-call grading) |
| `evals-test-suite.json` | Pre-deployment test scenarios (6 conversation simulations) |
| `../scripts/vapi/deploy-vapi.sh` | Deployment script — tries Vapi CLI, falls back to HTTP PATCH |
| `../scripts/vapi/push-assistant-and-sync-tools.sh` | **PATCH assistant + sync all tools’ server URL + Bearer** (API-stable) |
| `../scripts/vapi/sync-tool-auth.sh` | PATCH each function tool’s `server` + `Authorization` header |
| `../scripts/vapi/run-vapi-eval.sh` | Evals runner — executes test suite against live assistant |

## Quick Deploy

```bash
# 1. Install/auth once
supabase login
vapi login

# 2. One-command bridge: Supabase webhook -> Vapi assistant + tool auth
../scripts/vapi/connect-supabase-vapi.sh

# 3. (Optional) deploy eval criteria and run test suite
../scripts/vapi/deploy-vapi.sh --evals
../scripts/vapi/run-vapi-eval.sh
```

`connect-supabase-vapi.sh` does all of the following:
- deploys `webhook` to your linked Supabase project
- sets assistant `serverUrl` to `https://<project-ref>.supabase.co/functions/v1/webhook`
- pushes `riley-assistant.json`
- syncs function-tool bearer auth via `TOOL_API_KEY`

### Official CLI reference

- Docs: [Vapi CLI](https://docs.vapi.ai/cli) — `vapi assistant`, `vapi tool`, `vapi workflow`, `vapi auth`, etc.
- **There is no `vapi swarm` command.** Multi-agent / handoffs are usually **Squad** (Dashboard) or **`vapi workflow`**.

### CLI-only deploy (when your CLI version works)

Per docs, updates use **`--file`**, not `--config`:

```bash
vapi assistant update 450435e9-4562-4ddd-8429-54584d3285a7 --file riley-assistant.json
```

Some CLI releases **panic** on `assistant update` or fail **`vapi tool list`** when the API returns newer tool types (e.g. `handoff`). If that happens, use the script below (same result as the API).

### Reliable: push assistant + wire every tool’s `server` + Bearer

```bash
../scripts/vapi/push-assistant-and-sync-tools.sh
```

This **PATCH**es the assistant from `riley-assistant.json` (with SMS `from` merged from `.env`) and runs **`sync-tool-auth.sh`** so each function tool points at your **`serverUrl`** and **`Authorization: Bearer <TOOL_API_KEY>`**.

### cURL Fallback (no CLI)

```bash
export VAPI_API_KEY="your-key-here"

curl -X PATCH "https://api.vapi.ai/assistant/450435e9-4562-4ddd-8429-54584d3285a7" \
  -H "Authorization: Bearer $VAPI_API_KEY" \
  -H "Content-Type: application/json" \
  -d @riley-assistant.json
```

## Checking Results

```bash
# List recent calls
vapi call list

# Get call details (includes analysis/evaluation)
vapi call get <call-id>
# Look for "analysis.successEvaluation" in output

# Get current live config
../scripts/vapi/deploy-vapi.sh --get
```

## Modules

The system is organized into three modules, each with dedicated backend tool endpoints:

### Module 1: Language Detection

| Tool | Purpose | When |
|------|---------|------|
| `detect_language` | Detect English vs Chinese from first sentence, generate first response | On caller's first utterance |

### Module 2: Intent Classification

| Tool | Purpose | When |
|------|---------|------|
| `classify_intent` | Classify caller intent (booking, cancel, reschedule, questions, emergency) with confidence scoring and entity extraction | After language detection, on each utterance |

### Module 3: Appointment Booking Flow

| Tool | Purpose | When |
|------|---------|------|
| `get_current_date` | Get today's date in Toronto timezone | At start of booking |
| `parse_date` | Convert natural language dates to YYYY-MM-DD | Before check_availability |
| `get_next_available_dates` | Suggest next 5 open dates | When caller asks "earliest" |
| `check_availability` | Query PostgreSQL for open slots | Before offering any times |
| `validate_patient_info` | Verify name + phone (or email fallback) collected | Before booking |
| `is_booking_complete` | Check all required fields collected | Before booking |
| `get_booking_step` | Determine next step in booking flow | During booking |
| `fill_booking_fields` | Extract fields from patient utterance | On each patient response |
| `get_confirm_message` | Generate confirmation read-back message | Before calling book_appointment |
| `book_appointment` | Save confirmed booking to PostgreSQL | After patient confirms all details |
| `send_booking_confirmation` | Send SMS/email confirmation | Immediately after booking |
| `cancel_appointment` | Cancel existing appointment | When patient requests cancellation |

### Clinic Info & Patient Lookup

| Tool | Purpose | When |
|------|---------|------|
| `get_clinic_info` | Get hours, services, location, insurance info | When patient asks questions |
| `lookup_patient` | Check if patient exists in system | At start of booking/cancellation |

## Call Flow

```
Greeting → Module 1 Language Routing (lang_code + first_response) → Module 2 Intent Classification
  ├─ Booking → Service → Dentist → get_current_date → parse_date → check_availability
  │            → Offer times → Collect Name → Collect Phone (or email fallback)
  │            → validate_patient_info → is_booking_complete → Confirm → book_appointment
  │            → send_booking_confirmation → Wrap up
  ├─ Reschedule → Find appointment → check_availability → Confirm new time
  │              → book_appointment → send_booking_confirmation
  ├─ Cancel → Find appointment → cancel_appointment → Offer rebooking
  ├─ Question → get_clinic_info
  └─ Emergency → Direct to Southlake ER or same-day slot
```

## Assistant Details

- **ID**: `450435e9-4562-4ddd-8429-54584d3285a7`
- **Voice**: Elliot (Vapi native)
- **Model**: GPT-4o (temperature 0.3)
- **Transcriber**: Deepgram Nova-3 (multi-language)
- **Smart Endpointing**: LiveKit
- **Server URL**: `https://dental-ai.taildd3965.ts.net/functions/v1/webhook` (Tailscale tsnet)

## Evaluation

See [engineering-notebook/15-evaluation-technical-challenges.md](../engineering-notebook/15-evaluation-technical-challenges.md) for:
- Root-cause analysis of the 3 technical challenges
- Fix checklist and deployment pipeline
- Post-call evaluation rubric (analysis plan)
- Pre-deployment test suite (evals)
