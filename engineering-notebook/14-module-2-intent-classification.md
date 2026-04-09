# Page 14 — Module 2: Intent Classification

---

## 14.1 Overview

Module 2 receives the output from **Module 1** (`lang_code` + `first_response`) and performs **intent classification** on the patient's utterance. It determines what the patient wants to do (e.g., book an appointment, cancel, reschedule, ask a question) and routes the conversation to the appropriate handler.

> **Status:** Placeholder — details to be defined.

---

## 14.2 Input

Receives the output from Module 1:

```json
{
  "lang_code": "zh",
  "first_response": "您好！欢迎来到微笑牙科。请问今天有什么可以帮您的？"
}
```

Plus the patient's full first utterance for intent analysis.

---

## 14.3 Planned Intent Categories

| Intent | Description | Example |
|--------|-------------|---------|
| `book_appointment` | Patient wants to schedule a new appointment | "I'd like to book a cleaning" |
| `cancel_appointment` | Patient wants to cancel an existing appointment | "I need to cancel my appointment" |
| `reschedule_appointment` | Patient wants to change appointment time | "Can I move my appointment to Friday?" |
| `ask_hours` | Patient asks about clinic hours | "What are your hours?" |
| `ask_location` | Patient asks about clinic location | "Where are you located?" |
| `ask_services` | Patient asks about services offered | "Do you do root canals?" |
| `emergency` | Patient has a dental emergency | "I have a toothache, it's urgent" |
| `greeting` | Simple greeting, no specific intent | "Hello" |
| `unclear` | Intent cannot be determined | [ambiguous input] |

---

## 14.4 Planned Output

```json
{
  "module": "intent_classification",
  "intent": "book_appointment",
  "confidence": 0.92,
  "lang_code": "zh",
  "entities": {
    "service": "cleaning",
    "preferred_date": null,
    "preferred_time": null
  },
  "next_module": "appointment_booking"
}
```

---

## 14.5 Design Considerations

- Will use LLM-based classification (GPT-4o) with few-shot examples
- Must support both English and Chinese intent detection
- Should extract relevant entities (service type, date/time preferences)
- Confidence threshold determines whether to ask for clarification

---

*Previous: [Page 13 — Module 1: Language Detection](13-module-1-language-detection.md)*
*Next: [Page 15 — Module 3: Appointment Booking](15-module-3-appointment-booking.md)*
