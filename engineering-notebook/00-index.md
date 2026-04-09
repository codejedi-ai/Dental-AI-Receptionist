# 🦷 Engineering Notebook — Smile Dental AI Phone Dialer System

---

## Table of Contents

| Page | Document | Description |
|------|----------|-------------|
| 00 | `00-index.md` | This page — table of contents and project overview |
| 01 | `01-system-architecture.md` | Full system architecture, component diagram, data flow |
| 02 | `02-android-app-design.md` | Android app design: screens, UI, activities, layouts |
| 03 | `03-audio-pipeline.md` | AudioRecord → WebSocket → AudioTrack pipeline design |
| 04 | `04-websocket-protocol.md` | WebSocket message protocol, frame formats, state machine |
| 05 | `05-backend-server-design.md` | Node.js backend: WS server, call router, auth, agent pool |
| 06 | `06-agent-stub-design.md` | Test agent: loopback, registration, echo behavior |
| 07 | `07-api-compatibility.md` | Android 4.1+ (API 16) compatibility: constraints, workarounds |
| 08 | `08-security-design.md` | Auth tokens, network security, permission model |
| 09 | `09-build-deployment.md` | Gradle config, APK build, Node.js server setup, run instructions |
| 10 | `10-testing-strategy.md` | Testing plan: unit, integration, end-to-end, manual test matrix |
| 11 | `11-error-handling.md` | Error states, recovery, timeout handling, graceful degradation |
| 12 | `12-future-roadmap.md` | LiveKit migration, push-to-talk, call history, multi-device |
| 13 | `13-module-1-language-detection.md` | Module 1: Language detection, first response generation (zh/en) |
| 14 | `14-module-2-intent-classification.md` | Module 2: Intent classification with confidence scoring |
| 15 | `15-evaluation-technical-challenges.md` | Vapi Evaluation: 3 challenges, root-cause analysis, CLI deploy, analysis plan, test suites |
| 16 | `16-module-3-appointment-booking.md` | Module 3: Appointment booking state machine and flow |

---

## Project Summary

**Smile Dental AI Phone Dialer** is a native Android application that simulates an internal office phone system (PBX/VoIP). Users dial extension numbers (e.g. `101`, `102`) to reach AI agents on the local network. Audio is streamed in real-time via WebSocket using raw PCM 16-bit mono at 16kHz.

### Key Constraints
- **minSdkVersion 16** (Android 4.1 Jelly Bean) — non-negotiable
- **Pure Java + XML** — no Kotlin, no Jetpack Compose
- **Single external dependency:** OkHttp 3.12.x (last version supporting API 16)
- **Local network only** — no cloud dependencies, no authentication
- **Real-time audio** — bidirectional PCM streaming over WebSocket
- **No auth** — open access on local network (frontend only for now)

### Tech Stack
| Component | Technology |
|-----------|-----------|
| Android App | Native Java, XML layouts, Android SDK |
| WebSocket Client | OkHttp 3.12.13 |
| Audio I/O | AudioRecord + AudioTrack (Android SDK) |
| Backend Server | Node.js 14+, `ws` library, Express |
| Agent Stub | Node.js, `ws` library |
| Audio Format | PCM 16-bit, mono, 16kHz |

### System Architecture (High-Level)

```
┌─────────────────────────────────┐
│         Android Phone App       │
│    (Native Java, Android 4.1+)  │
│    ┌─────────────────────────┐  │
│    │  Phone Dialer UI        │  │
│    │  • Keypad (0-9, *, #)   │  │
│    │  • Call/Hang-up buttons  │  │
│    │  • Call timer            │  │
│    │  • Audio waveform        │  │
│    └──────────┬──────────────┘  │
│               │ WebSocket       │
└───────────────┼─────────────────┘
                │  ws://<local-ip>:3000
                │  ↕ JSON control msgs
                │  ↕ PCM audio binary
┌───────────────┼─────────────────┐
│          Backend Server         │
│         (Node.js + ws)          │
│                                 │
│  ┌──────────┐  ┌─────────────┐  │
│  │ Phonebook│  │ Call Router  │  │
│  │ 101→Agent│  │ mobile↔agent│  │
│  │ 102→Agent│  │ audio bridge│  │
│  └──────────┘  └──────┬──────┘  │
│                       │         │
│              ┌────────┴──────┐   │
│              │Agent WS Pool  │   │
│              │ws://agents    │   │
│              └───────────────┘   │
└─────────────────────────────────┘
                │
      ┌─────────┼──────────┐
      ↓         ↓          ↓
┌──────────┐ ┌──────────┐ ┌──────────┐
│ Agent 101│ │ Agent 102│ │ Agent 103│
│ Dental   │ │ Billing  │ │ Reminder │
│ Recept.  │ │          │ │          │
└──────────┘ └──────────┘ └──────────┘
```

---

## Quick Start

```bash
# 1. Start backend
cd backend && npm install && node server.js

# 2. Start test agent
cd agent-stub && npm install && node agent.js

# 3. Build & run Android app
# Open android-app/ in Android Studio → Run
# Or: gradlew assembleDebug → adb install app-debug.apk

# 4. Dial 101 from the app
```

---

*Each subsequent page in this notebook contains detailed design specifications, code structures, and implementation notes.*
