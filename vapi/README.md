# Vapi Configuration

This folder contains the Vapi assistant configuration for Riley, the Smile Dental Clinic AI receptionist.

## Files

- `riley-assistant.json` — The full assistant config (prompt, tools, voice, transcriber). This is the source of truth. Push updates to Vapi with:
  ```bash
  curl -X PATCH "https://api.vapi.ai/assistant/450435e9-4562-4ddd-8429-54584d3285a7" \
    -H "Authorization: Bearer $VAPI_API_KEY" \
    -H "Content-Type: application/json" \
    -d @vapi/riley-assistant.json
  ```

- `riley-current.json` — Snapshot of the live config from Vapi (for reference/diff).

## Tools

Riley has 4 tools that are called via webhook:

| Tool | Purpose | When |
|------|---------|------|
| `check_availability` | Look up open time slots | Before offering appointment times |
| `book_appointment` | Save a confirmed booking | After patient confirms all details |
| `cancel_appointment` | Cancel an existing appointment | When patient requests cancellation |
| `send_confirmation_email` | Email booking confirmation | After every successful booking |

## Call Flow

```
Greeting → Determine Intent
  ├─ Booking → Service → Dentist → New/Returning → Name → Phone → Date
  │            → check_availability → Offer times → Confirm
  │            → book_appointment → send_confirmation_email → Wrap up
  ├─ Reschedule → Find appointment → check_availability → Confirm new time
  │              → book_appointment → send_confirmation_email
  ├─ Cancel → Find appointment → cancel_appointment → Offer rebooking
  ├─ Question → Answer from knowledge base
  └─ Emergency → Direct to Southlake ER or same-day slot
```

## Assistant Details

- **ID**: `450435e9-4562-4ddd-8429-54584d3285a7`
- **Voice**: Elliot (Vapi native)
- **Model**: GPT-4o (temperature 0.5)
- **Transcriber**: Deepgram Nova-3
- **Smart Endpointing**: LiveKit
