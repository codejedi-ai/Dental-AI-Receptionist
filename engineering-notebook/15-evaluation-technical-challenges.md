# Page 15 — Vapi Platform Evaluation: Technical Challenges & Fixes

---

## 15.1 Overview

This page documents the **three technical challenges** in the Riley Dental AI Receptionist, provides root-cause analysis across the full stack (Vapi config → Go backend → PostgreSQL), defines concrete fixes, and includes a complete Vapi evaluation setup with CLI deployment, analysis plans, and test suites.

### Challenges Summary

| # | Challenge | Severity | Root Cause |
|---|-----------|----------|------------|
| 1 | AI not consistent collecting patient info — asks for email, address, DOB | 🔴 Critical | Conflicting system prompts across 3 configs; `book_appointment` params allow optional phone; `lookup_patient` returns extra fields |
| 2 | Missing services: Bridge, Consultation, Crown | 🟡 Medium | Service keywords incomplete in `intent_classifier.go` + `booking.go`; prompt service list stale |
| 3 | AI cannot query scheduling database for availability | 🔴 Critical | `check_availability` tool exists but may fail silently if: (a) Tailscale tsnet unreachable, (b) PostgreSQL down, (c) dentist table empty, (d) tool routing broken |

---

## 15.2 Challenge 1: Patient Info Collection — Root Cause & Fix

### Problem

The AI inconsistently collects patient information. It should strictly require: **full name + cell phone only**. Email is fallback only if no mobile. It must NOT ask for: home address, DOB, insurance.

### Root Cause Analysis (Full Stack)

#### Layer 1: Vapi Assistant Config (System Prompt)

Three conflicting configs exist:

| File | Status | Issue |
|------|--------|-------|
| `vapi/riley-current.json` | Live snapshot | Has strict rules BUT also mentions "confirm date of birth" in conversation flow — contradicts itself |
| `vapi/riley-assistant.json` | Deploy source | Has Module 1+2 logic but lacks explicit negative constraints (what NOT to collect) |
| `vapi-update.json` | Patch config | Has strict rules, but only gets deployed via manual curl — not part of main workflow |

**The AI sees whichever config was last pushed to Vapi.** If the live config contains contradictory instructions ("NO EMAIL" vs "confirm date of birth"), the LLM becomes inconsistent.

#### Layer 2: Tool Definitions

**`book_appointment` in `riley-assistant.json`:**
```json
"patientPhone": { "type": "string", "description": "Patient's cell phone..." },
// patientPhone is NOT in "required" array → LLM may skip it
```

**`book_appointment` in `vapi-update.json`:**
```json
"required": ["patientName", "patientPhone", "date", "time", "dentist", "service"]
// patientPhone IS required → stricter, but this config isn't deployed
```

#### Layer 3: Backend Tool Responses

**`LookupPatient` in `tools.go`** returns email and address:
```go
// BEFORE:
info += fmt.Sprintf(", Address: %s", *pt.Address)  // ← Gives LLM the idea to ask
info += fmt.Sprintf(", Email: %s", *pt.Email)       // ← Gives LLM the idea to ask
```

### Fixes Applied

#### Fix 1a: Unified System Prompt (`vapi/riley-assistant.json`)

New prompt includes explicit negative constraints:

```
=== STRICT PATIENT INFO RULES (MUST FOLLOW — NO EXCEPTIONS) ===

You MUST collect ONLY these fields:
  1. Full name (REQUIRED)
  2. Cell phone number (REQUIRED)

DO NOT collect any of these:
  ❌ Home address
  ❌ Date of birth
  ❌ Insurance details
  ❌ Emergency contact

EMAIL FALLBACK RULE:
  Only if patient explicitly says they do NOT have a mobile/cell phone,
  then and only then ask for email as backup for confirmation.
  Never ask for email first. Never ask for both phone and email.
```

#### Fix 1b: Hardened `book_appointment` Tool Definition

```json
"required": ["patientName", "date", "time", "dentist", "service"],
// phone is not "required" at JSON schema level because email fallback is valid,
// but the system prompt + validate_patient_info tool enforces: name+phone OR name+email
```

#### Fix 1c: Stripped `LookupPatient` Response

**File:** `vapi-backend/internal/tools/tools.go`

```go
// BEFORE:
info := fmt.Sprintf("Returning patient found — Name: %s, Phone: %s, UUID: %s", pt.Name, pt.Phone, pt.UUID)
if pt.Address != nil && *pt.Address != "" { info += fmt.Sprintf(", Address: %s", *pt.Address) }
if pt.Email != nil && *pt.Email != "" { info += fmt.Sprintf(", Email: %s", *pt.Email) }

// AFTER:
info := fmt.Sprintf("Returning patient found — Name: %s, Phone: %s", pt.Name, pt.Phone)
```

#### Fix 1d: Deploy Unified Config

```bash
cd vapi
./deploy-vapi.sh           # Deploy assistant config
./deploy-vapi.sh --evals   # Deploy analysis plan (evals)
```

### Evaluation Criteria

| Test | Expected |
|------|----------|
| Patient says "I want to book" | AI asks for name → then phone → does NOT ask for email/address/DOB |
| Patient says "I don't have a cell phone" | AI asks for email as fallback ONLY |
| `lookup_patient` tool call | Returns only name + phone |
| System prompt review | Contains "DO NOT collect: home address, DOB, insurance" |

---

## 15.3 Challenge 2: Missing Services — Root Cause & Fix

### Problem

The AI doesn't recognize **Bridge**, **Consultation**, or **Crown** as valid services.

### Root Cause Analysis

| File | Issue | Status |
|------|-------|--------|
| `vapi/riley-assistant.json` | Services list: `"Consultation, Cleaning, Filling, Bridge, Crown, Root Canal, Extraction, Whitening"` | ✅ Already present in deploy config |
| `vapi/riley-current.json` (live) | Services in prompt: `"Consultation (30-45min), Cleaning (30-45min), Filling, Bridge, Crown..."` | ✅ Already present |
| `intent_classifier.go` | Has `"bridge"`, `"consultation"`, `"crown"` in `serviceKeywords` map | ✅ Already present |
| `booking.go` | Has `"bridge"`, `"consultation"`, `"crown"` in `serviceMap` | ✅ Already present |

**The code and configs already have Bridge, Consultation, and Crown.** The issue is likely that the **live Vapi config** (`riley-current.json`) may be out of sync with the deploy source, or the LLM is not being prompted strongly enough about the available services.

### Fixes Applied

#### Fix 2a: Service List in System Prompt

Added to both `riley-assistant.json` and `vapi-update.json`:
```
Services: Consultation, Cleaning, Filling, Bridge, Crown, Root Canal, Extraction, Whitening, Implant, Invisalign, Pediatric, Emergency
```

#### Fix 2b: Service Keywords in Go Code

Already present. Verified in:
- `vapi-backend/internal/modules/intentclassifier/intent_classifier.go` — `serviceKeywords` map
- `vapi-backend/internal/modules/appointmentbooking/booking.go` — `serviceMap`
- Chinese keywords: `牙桥` (Bridge), `咨询` (Consultation), `牙冠` (Crown)

### Evaluation Criteria

| Test | Expected |
|------|----------|
| "I need a bridge appointment" | Service extracted as "Bridge" |
| "我想做咨询" (Chinese) | Service extracted as "Consultation" |
| "I need a crown" | Service extracted as "Crown" |
| AI lists services unprompted | Includes Bridge, Consultation, Crown |

---

## 15.4 Challenge 3: Scheduling Database Unreachable — Root Cause & Fix

### Problem

The `check_availability` tool cannot query the PostgreSQL scheduling database, blocking the entire booking workflow.

### Root Cause Analysis (Full Stack Trace)

```
Patient → Vapi (GPT-4o) → serverUrl → Tailscale tsnet → Go Handler → PostgreSQL
                                    ↓
                           https://dental-ai.taildd3965.ts.net/api/tools
```

#### Potential Failure Points

| # | Component | Check | Likely? |
|---|-----------|-------|---------|
| A | **Vapi → serverUrl connectivity** | Can Vapi reach `https://dental-ai.taildd3965.ts.net/api/tools`? | 🔴 **Most likely** — Tailscale auth key may have expired |
| B | **Tailscale tsnet server** | Is the `dental-vapi` container running and joined to tailnet? | 🔴 Very likely |
| C | **PostgreSQL container** | Is `dental-postgres` running and accepting connections? | 🟡 Possible |
| D | **Dentist seed data** | Does `SELECT * FROM dentists` return 3 rows? | 🟡 Possible |
| E | **Tool routing in handler.go** | Is `check_availability` case present in switch? | 🟢 Verified — present |
| F | **GetAllDentists returns empty** | If dentist table empty, returns generic error | 🟡 Possible |

#### Code Path Analysis

**1. Vapi sends tool call:**
```json
{"message":{"type":"tool-calls","tool_calls":[{"function":{"name":"check_availability","arguments":"{\"date\":\"2026-04-15\"}"}}]}}
```

**2. Handler routes** (`handler.go:104`):
```go
case "check_availability":
    result, status = tools.CheckAvailability(ctx, h.pg, call.Function.Arguments)
```

**3. Implementation** (`tools.go:23`):
```go
func CheckAvailability(ctx context.Context, pg *db.Postgres, args json.RawMessage) (string, string) {
    // ...
    allDentists, err := pg.GetAllDentists(ctx)
    if err != nil || len(allDentists) == 0 {
        return "Sorry, I'm having trouble accessing the dentist list right now.", "error"
        // ← THIS is the error the AI likely returns
    }
    // ...
}
```

**4. DB Query** (`postgres.go`):
```go
func (p *Postgres) GetAllDentists(ctx context.Context) ([]string, error) {
    rows, err := p.pool.Query(ctx, "SELECT name FROM dentists WHERE is_active = true ORDER BY name")
    // If pool is disconnected → err != nil → empty list → error message
}
```

### Fixes Applied

#### Fix 3a: Startup Health Check in `main.go`

Added dentist table verification at server startup:

```go
// After connecting, verify dentists table has data
dentists, err := pg.GetAllDentists(ctx)
if err != nil || len(dentists) == 0 {
    log.Printf("⚠️  WARNING: Dentist table is empty or unreachable. check_availability will fail.")
} else {
    log.Printf("✅ Dentists loaded: %v", dentists)
}
```

#### Fix 3b: Tool Execution Logging

Added detailed logging for `check_availability`:

```go
case "check_availability":
    log.Printf("📅 check_availability args: %s", string(call.Function.Arguments))
    result, status = tools.CheckAvailability(ctx, h.pg, call.Function.Arguments)
    log.Printf("📅 check_availability result: %s (status=%s)", result, status)
```

#### Fix 3c: `get_current_date` Tool Present

Verified `get_current_date` is in both the Vapi config tool definitions AND the handler switch statement. The AI needs this to resolve relative dates before calling `check_availability`.

### Diagnostic Commands

```bash
# 1. Check all containers running
docker compose ps

# 2. Check PostgreSQL health
docker exec dental-postgres pg_isready -U dental

# 3. Verify dentist seed data
docker exec dental-postgres psql -U dental -d dental -c "SELECT * FROM dentists;"

# 4. Check Vapi backend logs for tool execution
docker logs dental-vapi --tail 100 | grep "check_availability"

# 5. Test tool endpoint directly
curl -s -X POST https://dental-ai.taildd3965.ts.net/api/tools \
  -H "Content-Type: application/json" \
  -d '{"message":{"type":"tool-calls","tool_calls":[{"id":"t1","function":{"name":"check_availability","arguments":"{\"date\":\"2026-04-15\"}"}}]}}'

# 6. Check Tailscale connectivity
docker exec dental-vapi tailscale status
```

### Evaluation Criteria

| Test | Expected |
|------|----------|
| `GET /health` | Returns 200 |
| `POST /api/tools` with `check_availability` | Returns "Available slots on..." or "No available slots..." (NOT "trouble accessing") |
| `check_availability` for past date | Returns "That date is in the past" |
| `check_availability` for Sunday | Returns "clinic is closed on Sundays" |
| `docker logs dental-vapi` | Shows `✅ Dentists loaded: [...]` at startup |
| Vapi call transcript | AI says "Let me check availability" → tool fires → slots returned |

---

## 15.5 Vapi Platform Evaluation Setup

### 15.5.1 Deploy Unified Config via Vapi CLI

```bash
# Install CLI (one-time)
curl -sSL https://vapi.ai/install.sh | bash
vapi login

# Deploy the hardened assistant config
cd vapi
./deploy-vapi.sh

# Deploy analysis plan (call evaluation criteria)
./deploy-vapi.sh --evals

# Verify deployment
./deploy-vapi.sh --get
```

### 15.5.2 Call Analysis Plan (Post-Call Evaluation)

The `analysis-plan.json` configures Vapi to automatically grade every call after it ends:

```json
{
  "analysisPlan": {
    "successEvaluationPrompt": "Evaluate this dental appointment booking call...",
    "successEvaluationRubric": "Checklist"
  }
}
```

**Checklist items the AI judge evaluates:**

| # | Criterion | Pass Condition |
|---|-----------|----------------|
| 1 | Patient Info | Collected name + phone; did NOT ask for address/DOB/insurance |
| 2 | Service Identified | Valid service from allowed list |
| 3 | Availability Checked | `check_availability` called before offering times |
| 4 | Appointment Booked | Date, time, dentist confirmed |
| 5 | Confirmation Sent | SMS/email confirmation mentioned |
| 6 | Language Consistency | Same language as patient |
| 7 | One Question at a Time | No multi-question prompts |

### 15.5.3 Pre-Deployment Evals (Test Suites)

The `evals-test-suite.json` defines 6 simulated conversation scenarios:

| Test | Purpose |
|------|---------|
| Booking Flow — English, Cleaning | Full booking with name+phone only |
| Booking Flow — Chinese, Crown | Chinese language booking |
| No Mobile — Email Fallback | Email asked only after patient says no phone |
| Service Recognition — Bridge | "Bridge" recognized as valid service |
| Availability Check | `check_availability` called before offering times |
| Rejection — No Address | Agent never asks for address |

### 15.5.4 Running Evals

```bash
# Run all evals
./run-vapi-eval.sh

# Run specific test
./run-vapi-eval.sh --test "Bridge"

# Check call results via CLI
vapi call list
vapi call get <call-id>
# Look for "analysis.successEvaluation" in output
```

### 15.5.5 Automated Backend Tests

```bash
# From engineering-notebook/
./run-evaluation.sh
```

This tests the backend tool endpoints directly (no Vapi needed):
- Health check
- Tool endpoint POST
- `check_availability` (future date, past date, Sunday)
- `book_appointment` (missing phone → rejected)
- Language detection (zh/en)
- Intent classification

---

## 15.6 Fix Checklist

### Challenge 1: Patient Info Collection
- [x] Unified system prompt in `riley-assistant.json` with strict negative constraints
- [x] Removed email/address from `LookupPatient` response in `tools.go`
- [ ] Deploy updated config: `./deploy-vapi.sh`
- [ ] Verify: AI does NOT ask for email/address/DOB in test calls

### Challenge 2: Missing Services
- [x] Verified Bridge, Consultation, Crown in system prompt
- [x] Verified service keywords in `intent_classifier.go` and `booking.go`
- [x] Verified Chinese keywords (牙桥, 咨询, 牙冠)
- [ ] Deploy updated config: `./deploy-vapi.sh`
- [ ] Test: "bridge" and "consultation" recognized in both languages

### Challenge 3: Availability Query
- [x] Added startup dentist table verification in `main.go`
- [x] Added tool execution logging for `check_availability`
- [x] Verified `get_current_date` in config and handler
- [ ] Verify Tailscale connectivity: `docker exec dental-vapi tailscale status`
- [ ] Verify PostgreSQL: `docker exec dental-postgres psql -U dental -d dental -c "SELECT * FROM dentists;"`
- [ ] Test: `check_availability` returns slots (not error)
- [ ] Deploy updated config: `./deploy-vapi.sh`

### Deploy Pipeline
- [ ] `cd vapi && ./deploy-vapi.sh`
- [ ] `cd vapi && ./deploy-vapi.sh --evals`
- [ ] `cd engineering-notebook && ./run-evaluation.sh`
- [ ] `cd vapi && ./run-vapi-eval.sh`
- [ ] Make a test phone call and verify analysis results: `vapi call get <id>`

---

## 15.7 Files Reference

| File | Purpose |
|------|---------|
| `vapi/riley-assistant.json` | Hardened assistant config (deploy source) |
| `vapi/analysis-plan.json` | Call analysis evaluation rubric |
| `vapi/evals-test-suite.json` | Pre-deployment test scenarios |
| `vapi/deploy-vapi.sh` | CLI deployment script |
| `vapi/run-vapi-eval.sh` | Evals runner |
| `vapi-backend/internal/tools/tools.go` | Backend tool implementations |
| `vapi-backend/internal/handlers/handler.go` | Tool call routing |
| `vapi-backend/cmd/main.go` | Server startup + health checks |
| `engineering-notebook/run-evaluation.sh` | Backend API test suite |

---

*Previous: [Page 14 — Module 2: Intent Classification](14-module-2-intent-classification.md)*
*Next: [Page 16 — Module 3: Appointment Booking](16-module-3-appointment-booking.md)*
