# Page 12 — Future Roadmap

---

## 12.1 Overview

This page outlines planned enhancements and potential future directions for the Smile Dental AI Phone Dialer system. Items are prioritized by impact and implementation complexity.

---

## 12.2 Phase 1: Production Readiness

### 1.2.1 TLS Encryption (wss://)

**Priority:** 🔴 Critical
**Effort:** Medium

Replace unencrypted WebSocket with TLS-encrypted WebSocket.

```
Current:  ws://192.168.1.100:3000/call
Future:   wss://192.168.1.100:3000/call
```

**Changes Required:**
- Backend: Add HTTPS/WSS server with SSL certificate
- Android: Update config.xml URL to wss://
- Certificate: Self-signed for local network, or internal CA

**Backend TLS Setup:**
```javascript
const https = require('https');
const fs = require('fs');

const server = https.createServer({
    key: fs.readFileSync('/etc/smiledental/private-key.pem'),
    cert: fs.readFileSync('/etc/smiledental/certificate.pem')
}, app);
```

**Android Trust Store:**
```java
// For self-signed certificates, add to trust store
// Or use OkHttp's certificate pinning
OkHttpClient client = new OkHttpClient.Builder()
    .certificatePinner(new CertificatePinner.Builder()
        .add("192.168.1.100", "sha256/expected-cert-hash")
        .build())
    .build();
```

---

### 1.2.2 Call History

**Priority:** 🟡 High
**Effort:** Low

Store and display recent calls on the dialer screen.

**Data Model:**
```java
public class CallRecord {
    String number;
    String agentName;
    long timestamp;      // When call occurred
    long duration;       // Call duration in ms
    CallOutcome outcome; // CONNECTED, MISSED, BUSY, ERROR
}
```

**Storage:** SQLite database or SharedPreferences (for small history)

**UI Enhancement:**
```
┌─────────────────────────────────────┐
│         Smile Dental 🦷             │
│                                     │
│  ┌───────────────────────────────┐  │
│  │  Recent Calls                 │  │
│  │  ───────────────────────────  │  │
│  │  📞 101  Smile Dental... 2m   │  │
│  │  📞 102  Billing Agent  5m    │  │
│  │  ❌ 105  Emergency      —     │  │
│  └───────────────────────────────┘  │
│                                     │
│  ┌───────────────────────────────┐  │
│  │                               │  │
│  │        (number display)       │  │
│  │                               │  │
│  └───────────────────────────────┘  │
│                                     │
│  [Keypad...]                        │
└─────────────────────────────────────┘
```

---

### 1.2.3 Contact Favorites

**Priority:** 🟡 High
**Effort:** Medium

Allow users to save frequently-dialed numbers as favorites.

**UI:** Quick-access buttons above keypad for favorite numbers.

```
┌─────────────────────────────────────┐
│  ⭐ 101    ⭐ 102    ⭐ 103         │  ← Favorite buttons
│                                     │
│  ┌───────────────────────────────┐  │
│  │                               │  │
│  │        (number display)       │  │
│  │                               │  │
│  └───────────────────────────────┘  │
│                                     │
│  [Keypad...]                        │
└─────────────────────────────────────┘
```

---

### 1.2.4 Rate Limiting

**Priority:** 🟡 High
**Effort:** Low

Prevent abuse by limiting call attempts per time window.

**Backend Implementation:**
```javascript
const callAttempts = new Map();  // IP -> [{timestamp}]

function checkRateLimit(clientIp) {
    const now = Date.now();
    const windowMs = 60000;  // 1 minute window
    const maxAttempts = 10;

    const attempts = callAttempts.get(clientIp) || [];
    const recent = attempts.filter(t => now - t < windowMs);

    if (recent.length >= maxAttempts) {
        return false;  // Rate limited
    }

    recent.push(now);
    callAttempts.set(clientIp, recent);
    return true;
}
```

---

## 12.3 Phase 2: Feature Enhancements

### 1.2.1 LiveKit Migration

**Priority:** 🔴 Critical (for production AI)
**Effort:** High

Replace custom WebSocket audio bridge with LiveKit for production-grade real-time communication.

**Why LiveKit?**
- SFU architecture for multi-party calls
- Built-in simulcast (quality adaptation)
- Native SDKs for Android
- Recording, transcription plugins
- Better NAT traversal

**Architecture Change:**
```
┌─────────────────────────────────────────────────────────────┐
│              Current Architecture                           │
│                                                             │
│  Mobile ──WS(PCM)──► Backend ──WS(PCM)──► Agent             │
│         ◄──────────          ◄──────────                    │
│                                                             │
│  Simple binary bridge, no codec, no multi-party             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              LiveKit Architecture                           │
│                                                             │
│  Mobile ──WebRTC──► LiveKit Server ──WebRTC──► Agent        │
│           (Opus)       (SFU)          (Opus)                │
│                                                             │
│  Production-grade, codec support, multi-party, plugins      │
└─────────────────────────────────────────────────────────────┘
```

**Migration Steps:**
1. Add LiveKit Android SDK dependency
2. Replace WebSocket audio pipeline with LiveKit room connection
3. Replace backend WS audio bridge with LiveKit server
4. Update agent stub to use LiveKit SDK
5. Add token generation endpoint on backend

**Android Changes:**
```java
// Current: WebSocket + AudioRecord
WebSocketManager wsManager = new WebSocketManager(...);
AudioEngine audioEngine = new AudioEngine(...);

// Future: LiveKit
Room liveKitRoom = new Room(context);
liveKitRoom.connect(serverUrl, token);

// Publish microphone
LocalAudioTrack audioTrack = LocalAudioTrack.create(context, true);
liveKitRoom.localParticipant.publishTrack(audioTrack);

// Subscribe to agent audio
liveKitRoom.addListener(new RoomListener() {
    @Override
    public void onTrackPublished(RemoteTrackPublication pub, RemoteParticipant participant) {
        pub.setSubscribed(true);
    }
});
```

---

### 1.2.2 Push Notifications

**Priority:** 🟢 Medium
**Effort:** High

Allow agents to initiate calls to the mobile device (incoming calls).

**Implementation:**
- Firebase Cloud Messaging (FCM) for push notifications
- Notification shows "Incoming call from [Agent]"
- Tap notification opens app and auto-connects

**Note:** Requires Google Play Services, not available on all Android 4.x devices.

---

### 1.2.3 Voicemail

**Priority:** 🟢 Medium
**Effort:** Medium

Leave a voicemail when agent is unavailable.

**Flow:**
```
User dials 101 → Agent unavailable → Prompt: "Leave a voicemail?"
    → User records message → Hang up → Agent receives voicemail notification
```

**Storage:** PCM audio file on backend, accessible via HTTP endpoint.

---

### 1.2.4 Call Transfer

**Priority:** 🟢 Medium
**Effort:** Medium

Allow agents to transfer calls to other agents.

**Protocol Extension:**
```json
// Agent → Server: Transfer request
{
    "action": "transfer",
    "toNumber": "102",
    "reason": "Billing inquiry"
}

// Server → Mobile: Transfer notification
{
    "event": "transferring",
    "toAgent": "Billing Agent",
    "reason": "Billing inquiry"
}

// Server → New Agent: Incoming transferred call
{
    "event": "incoming_transfer",
    "fromAgent": "Receptionist",
    "callerId": "user-token-123"
}
```

---

## 12.4 Phase 3: Advanced Features

### 1.2.1 AI Agent Integration

**Priority:** 🔴 Critical (for actual AI functionality)
**Effort:** Very High

Replace agent stub with real AI pipeline: STT → LLM → TTS.

```
┌─────────────────────────────────────────────────────────────┐
│              AI Agent Pipeline                              │
│                                                             │
│  PCM Audio In                                             │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                       │
│  │  Voice Activity │  ← Detect end of user speech          │
│  │  Detection (VAD)│                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │  Speech-to-Text │  ← Whisper, Google STT, etc.          │
│  │  (STT)          │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │  LLM Processing │  ← GPT, Claude, custom model          │
│  │                 │  ← System prompt: dental receptionist │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │  Text-to-Speech │  ← ElevenLabs, Google TTS, etc.       │
│  │  (TTS)          │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                               │
│  PCM Audio Out                                            │
│                                                             │
│  Target latency: < 1 second from end-of-speech to response  │
└─────────────────────────────────────────────────────────────┘
```

---

### 1.2.2 Multi-Device Support

**Priority:** 🟢 Medium
**Effort:** High

Allow multiple phones to call the same agent simultaneously.

**Changes:**
- Backend: Support multiple mobile connections per agent
- Audio mixing: Combine audio from multiple phones
- Agent: Handle multi-party conversation

---

### 1.2.3 Web Dashboard

**Priority:** 🟢 Medium
**Effort:** Medium

Admin web interface for managing agents, phonebook, and monitoring.

**Features:**
- View registered agents and their status
- Add/edit/remove phonebook entries
- View active calls
- View call history and statistics
- Manage auth tokens

**Tech:** Simple HTML/CSS/JS frontend, Express API backend.

---

### 1.2.4 Analytics & Monitoring

**Priority:** 🟡 High
**Effort:** Medium

Track system usage and health metrics.

**Metrics to Track:**
| Metric | Purpose |
|--------|---------|
| Calls per day | Usage trends |
| Average call duration | Engagement |
| Connection success rate | Reliability |
| Average latency | Quality |
| Error rate | System health |
| Most-dialed numbers | Agent popularity |

**Dashboard:**
```
┌─────────────────────────────────────────────────────────────┐
│              Smile Dental Analytics                         │
│                                                             │
│  Today's Stats:                                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │  Calls   │ │  Avg Dur │ │ Success  │ │  Errors  │      │
│  │   47     │ │  3m 12s  │ │  98.2%   │ │    2     │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
│                                                             │
│  Most Dialed:                                               │
│  101  ████████████████████  62%                            │
│  102  ████████              21%                            │
│  103  ████                  11%                            │
│  105  ██                     6%                            │
│                                                             │
│  Active Agents: 4/5                                         │
│  ✅ 101  Smile Dental Receptionist                          │
│  ✅ 102  Billing Agent                                      │
│  ✅ 103  Appointment Reminder                               │
│  ❌ 104  Patient Support (offline)                          │
│  ✅ 105  Emergency Line                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 12.5 Priority Matrix

```
┌─────────────────────────────────────────────────────────────┐
│              Implementation Priority Matrix                  │
│                                                             │
│  High Impact │  1. TLS Encryption        4. LiveKit         │
│              │  2. Call History          5. AI Integration   │
│              │  3. Rate Limiting         6. Multi-Device     │
│              │                                             │
│  Low Impact  │  7. Voicemail             8. Push Notifs      │
│              │  9. Call Transfer         10. Web Dashboard   │
│              │  11. Analytics            12. Favorites       │
│              │                                             │
│              ├──────────────────┬────────────────────────── │
│              │   Quick Wins     │   Major Projects          │
│              └──────────────────┴────────────────────────── │
│                                                             │
│  Start with: TLS, Call History, Rate Limiting               │
│  Then: LiveKit migration, AI agent integration              │
│  Later: Advanced features (voicemail, transfer, analytics)  │
└─────────────────────────────────────────────────────────────┘
```

---

## 12.6 Technical Debt

| Item | Impact | Effort to Fix |
|------|--------|---------------|
| Java 7 syntax constraints | Development speed | High (requires minSdk bump) |
| No automated UI tests | Regression risk | Medium |
| Hardcoded config values | Deployment friction | Low |
| No CI/CD pipeline | Release friction | Medium |
| Single-process backend | Scalability limit | High |
| In-memory agent registry | Data loss on restart | Low (add persistence) |

---

## 12.7 Version Roadmap

| Version | Features | Target Date |
|---------|----------|-------------|
| **1.0** | Core dialer, WS audio bridge, agent stub | ✅ Current |
| **1.1** | TLS, call history, rate limiting | — |
| **1.2** | LiveKit migration, multi-device | — |
| **2.0** | AI agent integration (STT→LLM→TTS) | — |
| **2.1** | Voicemail, call transfer, push notifications | — |
| **3.0** | Web dashboard, analytics, multi-tenant | — |

---

*← Back to [Page 00 — Index](00-index.md)*
*Next: [Page 13 — Module 1: Language Detection](13-module-1-language-detection.md)*
