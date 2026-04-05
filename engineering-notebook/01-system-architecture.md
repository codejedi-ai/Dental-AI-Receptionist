# Page 01 — System Architecture

---

## 1.1 Overview

The Smile Dental AI Phone Dialer is a **three-component system** running entirely on a local network:

1. **Android Phone App** — Native Java client with phone dialer UI and real-time audio streaming
2. **Backend Server** — Node.js WebSocket server handling authentication, call routing, and audio bridging
3. **AI Agents** — Separate WebSocket services that register with the backend and handle calls

```
┌──────────────────────────────────────────────────────────────────┐
│                     LOCAL NETWORK (192.168.x.x)                  │
│                                                                  │
│  ┌──────────────┐           ┌──────────────────────────────┐    │
│  │ Android Phone│           │     Backend (Node.js)        │    │
│  │ (API 16+)    │◄──WS────►│     0.0.0.0:3000             │    │
│  │              │  audio    │                              │    │
│  │ ┌──────────┐ │  + json   │  ┌────────┐  ┌───────────┐  │    │
│  │ │ Keypad   │ │           │  │  Auth  │  │Call Router │  │    │
│  │ │ Call UI  │ │           │  │Validate│  │mobile↔agent│  │    │
│  │ │ Audio I/O│ │           │  │ tokens │  │PCM bridge  │  │    │
│  │ └──────────┘ │           │  └────────┘  └─────┬─────┘  │    │
│  └──────────────┘           │                    │         │    │
│                             │         ┌──────────┼───────┐ │    │
│                             │         │  Agent   │  Pool │ │    │
│                             │         │  WS Map  ▼       │ │    │
│                             │         │ 101 → ws://..    │ │    │
│                             │         │ 102 → ws://..    │ │    │
│                             │         │ 103 → ws://..    │ │    │
│                             │         └──────────────────┘ │    │
│                             └──────────────┬───────────────┘    │
│                                            │ WS connections      │
│                    ┌───────────────────────┼──────────┐          │
│                    │                       │          │          │
│              ┌─────┴──────┐  ┌─────┴──────┐  ┌─────┴──────┐    │
│              │  Agent 101 │  │  Agent 102 │  │  Agent 103 │    │
│              │  Dental    │  │  Billing   │  │  Reminder  │    │
│              │  Recept.   │  │  Agent     │  │  Agent     │    │
│              │  :4001     │  │  :4002     │  │  :4003     │    │
│              └────────────┘  └────────────┘  └────────────┘    │
│                                                                  │
╚══════════════════════════════════════════════════════════════════╝
```

---

## 1.2 Component Responsibilities

### Android Phone App
| Responsibility | Implementation |
|---------------|----------------|
| User Interface | DialerActivity (keypad), CallActivity (call states) |
| Audio Capture | `AudioRecord` — PCM 16-bit, mono, 16kHz |
| Audio Playback | `AudioTrack` — streaming mode, same format |
| Network | OkHttp WebSocket client for control + binary audio frames |
| State Management | Finite state machine: DIALING → RINGING → CONNECTED → ENDED |
| Permissions | RECORD_AUDIO, INTERNET, MODIFY_AUDIO_SETTINGS |

### Backend Server (Node.js)
| Responsibility | Implementation |
|---------------|----------------|
| WebSocket Server | `ws` library — two upgrade paths: `/call` and `/agent` |
| Phonebook | Static JSON mapping numbers to agent names |
| Agent Registry | In-memory map: `number → WebSocket connection` |
| Call Routing | Match dialed number → find agent → establish audio bridge |
| Audio Bridge | Bidirectional binary frame forwarding between mobile ↔ agent |
| Lifecycle Management | Handle disconnects, timeouts, cleanup |

### AI Agents
| Responsibility | Implementation |
|---------------|----------------|
| Registration | Connect to `ws://localhost:3000/agent`, send register message |
| Call Handling | Accept incoming call notifications from backend |
| Audio Processing | Receive PCM frames → process (AI/TTS/STT) → send PCM response |
| Loopback Mode (stub) | Echo received audio back for testing |

---

## 1.3 Data Flow

### Call Initiation Flow
```
┌────────┐         ┌─────────┐         ┌────────┐
│ Mobile │         │ Backend │         │ Agent  │
└───┬────┘         └────┬────┘         └───┬────┘
    │                   │                  │
    │  WS Connect       │                  │
    │──────────────────►│                  │
    │                   │                  │
    │  {dial, "101"}    │                  │
    │──────────────────►│                  │
    │                   │  Lookup "101"    │
    │                   │  → Agent found   │
    │                   │                  │
    │                   │  {incoming_call} │
    │                   │─────────────────►│
    │                   │                  │
    │                   │  {accept}        │
    │                   │◄─────────────────│
    │                   │                  │
    │  {ringing}        │                  │
    │◄──────────────────│                  │
    │                   │                  │
    │  {connected}      │                  │
    │◄──────────────────│                  │
    │                   │                  │
    │  PCM binary       │  PCM binary      │
    │──────────────────►│─────────────────►│
    │                   │                  │
    │  PCM binary       │  PCM binary      │
    │◄──────────────────│◄─────────────────│
    │                   │                  │
    │  {hangup}         │  {ended}         │
    │──────────────────►│─────────────────►│
    │                   │                  │
    │  WS Close         │  Cleanup         │
    │◄──────────────────│◄─────────────────│
```

### Audio Bridge Flow
```
┌─────────────────────────────────────────────────────────────┐
│                    Audio Bridge (Backend)                    │
│                                                             │
│  Mobile WS                    Agent WS                      │
│  ┌────────┐                  ┌────────┐                     │
│  │  Read  │──binary frame──►│  Write │                     │
│  │Binary  │                  │Binary  │                     │
│  └────────┘                  └────────┘                     │
│       ▲                        │                            │
│       │                        │                            │
│  ┌────────┐                  ┌────────┐                     │
│  │  Write │◄──binary frame───│  Read  │                     │
│  │Binary  │                  │Binary  │                     │
│  └────────┘                  └────────┘                     │
│                                                             │
│  Both directions simultaneously (full-duplex)               │
│  No transcoding — raw PCM pass-through                      │
│  No buffering strategy — immediate forwarding               │
└─────────────────────────────────────────────────────────────┘
```

---

## 1.4 Network Topology

```
                    ┌─────────────────────┐
                    │   WiFi Router       │
                    │   192.168.1.1       │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
    ┌─────────┴──────┐ ┌──────┴───────┐ ┌──────┴───────┐
    │  Android Phone │ │  Backend     │ │  Agent Host  │
    │  192.168.1.50  │ │  192.168.1.100│ │  192.168.1.101│
    │  (WiFi)        │ │  :3000       │ │  :4001-4003  │
    └────────────────┘ └──────────────┘ └──────────────┘
```

All communication stays within the local network. No external internet access required after initial APK installation.

---

## 1.5 Design Decisions

| Decision | Rationale |
|----------|-----------|
| WebSocket over HTTP | Full-duplex, low-latency, binary frame support |
| PCM 16kHz mono | Telephony quality, low bandwidth (~32 KB/s), compatible with all Android versions |
| No codec | Simplicity, local network has sufficient bandwidth |
| Token-based auth | Stateless, simple, no session management needed |
| In-memory agent registry | Single-process backend, no persistence needed |
| OkHttp 3.12.x | Last version supporting API 16, stable, well-tested |
| No Kotlin/Compose | Maximum backward compatibility with Android 4.1 |

---

## 1.6 Failure Modes

| Failure | Detection | Recovery |
|---------|-----------|----------|
| Backend unreachable | WebSocket connection timeout | Show error toast, return to dialer |
| Agent not registered | Phonebook lookup fails | Show "Number not available" |
| Audio stream interrupted | WebSocket close/error event | Transition to ENDED state |
| Network drops | Read/write timeout (30s) | Auto-hangup, show "Connection lost" |
| Agent disconnects | Backend detects WS close | Send `ended` to mobile, cleanup |
| Mobile app crashes | Backend detects WS close | Send `ended` to agent, cleanup |

---

*Next: [Page 02 — Android App Design](02-android-app-design.md)*
