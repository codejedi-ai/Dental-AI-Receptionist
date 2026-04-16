# AI Tool Use Guide (Dental Reception)

This guide explains exactly how the AI should call each dental_action.

In this repo, each dental_action is one of the actions the AI can do with the tool, and the single tool the AI calls is dispatch_dental_action.

## 1) Primary Call Pattern (Router Tool)

In this project, the AI should call the single router tool:

- Tool name: `dispatch_dental_action`
- Required field: `operation`
- Optional fields: `payload`, `requestId`

Use this shape for every call:

```json
{
  "operation": "<one_of_supported_operations>",
  "payload": {},
  "requestId": "optional-idempotency-key"
}
```

Notes:
- `payload` can be `{}` for operations that need no inputs.
- `requestId` is optional but useful for retries/idempotency.

## 2) Supported Operations

The router supports these dental_action operations:

- `health_check`
- `get_current_date`
- `get_clinic_info`
- `get_dentists`
- `parse_date`
- `get_next_available_dates`
- `check_availability`
- `book_appointment`
- `cancel_appointment`
- `send_booking_confirmation`

## 3) Per-Function Call Examples

### `health_check`
Purpose: verify Supabase connectivity and minimal DB read.

```json
{
  "operation": "health_check",
  "payload": {}
}
```

### `get_current_date`
Purpose: get current Toronto date/time for relative scheduling.

```json
{
  "operation": "get_current_date",
  "payload": {}
}
```

### `get_clinic_info`
Purpose: return clinic info by topic.

Inputs:
- `topic` (optional): `general`, `hours`, `services`, `location`, `insurance`, `emergency`

```json
{
  "operation": "get_clinic_info",
  "payload": {
    "topic": "hours"
  }
}
```

### `get_dentists`
Purpose: list active dentists.

```json
{
  "operation": "get_dentists",
  "payload": {}
}
```

### `parse_date`
Purpose: convert natural language date into `YYYY-MM-DD`.

Inputs:
- `date_text` (preferred)
- `dateText` (alias)

```json
{
  "operation": "parse_date",
  "payload": {
    "date_text": "next Tuesday"
  }
}
```

### `get_next_available_dates`
Purpose: suggest upcoming dates with open slots.

Inputs:
- `days` (optional number): lookahead window (default 14, max 60)

```json
{
  "operation": "get_next_available_dates",
  "payload": {
    "days": 21
  }
}
```

### `check_availability`
Purpose: return available slots for a date.

Inputs:
- `date` (required): `YYYY-MM-DD`
- `dentist` (optional): full dentist name

```json
{
  "operation": "check_availability",
  "payload": {
    "date": "2026-04-22",
    "dentist": "Dr. Sarah Kim"
  }
}
```

Alias behavior:
- `provider` is accepted and normalized to `dentist`
- `provider: "any"` is normalized to empty dentist filter

### `book_appointment`
Purpose: create a confirmed appointment.

Required inputs:
- `patientName`
- `date` (`YYYY-MM-DD`)
- `time` (`HH:MM` 24h)
- `dentist`
- `service`

Optional inputs:
- `patientPhone`
- `patientEmail`
- `notes`

```json
{
  "operation": "book_appointment",
  "payload": {
    "patientName": "Alex Chen",
    "patientPhone": "+1-416-555-0134",
    "date": "2026-04-22",
    "time": "10:30",
    "dentist": "Dr. Sarah Kim",
    "service": "Cleaning",
    "notes": "Prefers morning appointments"
  }
}
```

Alias behavior:
- `phone` -> `patientPhone`
- `reason` -> `service`
- `provider` -> `dentist` (`provider: "any"` -> empty dentist filter)
- `timezone` is ignored and removed

### `cancel_appointment`
Purpose: cancel a confirmed appointment.

Required inputs:
- `patientName`
- `date` (`YYYY-MM-DD`)

```json
{
  "operation": "cancel_appointment",
  "payload": {
    "patientName": "Alex Chen",
    "date": "2026-04-22"
  }
}
```

### `send_booking_confirmation`
Purpose: acknowledge confirmation delivery message.

```json
{
  "operation": "send_booking_confirmation",
  "payload": {}
}
```

## 4) Recommended AI Call Sequence

For booking conversations, use this order:

1. `get_current_date` (if user gave relative dates)
2. `parse_date` (if input is natural language date)
3. `check_availability`
4. `book_appointment`
5. `send_booking_confirmation`

For cancellations:

1. `cancel_appointment`

For general questions:

- Clinic facts: `get_clinic_info`
- Dentist list: `get_dentists`
- Earliest dates: `get_next_available_dates`

## 5) Guardrails for the AI

- Always call `check_availability` before presenting specific booking times.
- Do not call `book_appointment` until required booking fields are collected.
- Never ask the patient to convert the date to `YYYY-MM-DD`; use `parse_date` on the date the patient said, then pass the normalized result to scheduling tools.
- If a tool returns an error/result asking for missing fields, ask the user only for those missing fields and retry.

## 6) Legacy Direct Handler Form (Internal)

Internally, handlers also exist as a flat map (`defaultToolHandlers`) where calls look like:

```ts
await defaultFunctions["check_availability"]({ date: "2026-04-22" });
```

For AI tool use, prefer the router (`dispatch_dental_action`) schema above.
