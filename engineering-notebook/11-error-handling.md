# Page 11 — Error Handling

---

## 11.1 Overview

Error handling is critical for a real-time audio application. Users must always understand what's happening and how to recover. This page documents every error scenario and the system's response.

---

## 11.2 Error Classification

### By Severity

| Level | Description | User Impact | Recovery |
|-------|------------|-------------|----------|
| **Critical** | System cannot function | Call impossible | User action required |
| **Major** | Active call affected | Call interrupted | Auto or user action |
| **Minor** | Degraded experience | Call continues | Auto-recovery |
| **Informational** | Non-impact | None | Logged only |

### By Source

| Source | Examples |
|--------|----------|
| Network | Connection refused, timeout, drop |
| Server | Auth failure, number not found, agent unavailable |
| Audio | Mic unavailable, speaker error, buffer underrun |
| Client | Permission denied, invalid input, crash |
| Agent | Agent crash, agent disconnect, agent busy |

---

## 11.3 Error Scenarios — Android App

### E1: Backend Server Unreachable

```
Trigger: WebSocket connection fails (server not running, wrong IP, firewall)
Detection: OkHttp onFailure callback after connect timeout (~10s)
Severity: Critical

User Experience:
┌─────────────────────────────────┐
│  Dialer Screen                  │
│                                 │
│  1 0 1                          │
│                                 │
│  ⚠️  "Cannot connect to server" │  ← Toast message
│     "Check network connection"  │
│                                 │
│  [Keypad...]                    │
│  [Call button]                  │
└─────────────────────────────────┘

Recovery:
1. Stay on dialer screen
2. User can retry by pressing call again
3. User should verify backend URL in config.xml
```

### E3: Number Not Found

```
Trigger: Dialed number not in phonebook.json
Detection: Server sends {event: "busy", reason: "Number not available"}
Severity: Major

User Experience:
┌─────────────────────────────────┐
│  Call Ended Screen              │
│                                 │
│       ❌                        │
│   Call Ended                    │
│                                 │
│   Number not available on       │
│   this network                  │
│                                 │
│   Duration: 0s                  │
│                                 │
│   [ Back to Dialer ]            │
└─────────────────────────────────┘

Recovery:
1. Show call ended screen with error reason
2. User presses "Back to Dialer" to try again
3. User should verify the number
```

### E4: Agent Not Available

```
Trigger: Number exists in phonebook but no agent registered
Detection: Server sends {event: "busy", reason: "Agent not available"}
Severity: Major

User Experience:
┌─────────────────────────────────┐
│  Call Ended Screen              │
│                                 │
│       ❌                        │
│   Call Ended                    │
│                                 │
│   Smile Dental Receptionist     │
│   is not currently available    │
│                                 │
│   Duration: 0s                  │
│                                 │
│   [ Call Again ]                │
│   [ Back to Dialer ]            │
└─────────────────────────────────┘

Recovery:
1. Show call ended screen
2. "Call Again" option available (agent may have reconnected)
3. "Back to Dialer" to try different number
```

### E5: Call Timeout (No Answer)

```
Trigger: Ringing state exceeds timeout (60 seconds)
Detection: Timer in CallActivity
Severity: Major

User Experience:
┌─────────────────────────────────┐
│  Call Ended Screen              │
│                                 │
│       ❌                        │
│   Call Ended                    │
│                                 │
│   No answer                     │
│                                 │
│   Duration: 0s                  │
│                                 │
│   [ Call Again ]                │
│   [ Back to Dialer ]            │
└─────────────────────────────────┘

Recovery:
1. Auto-transition to call ended screen
2. User can redial or return to dialer
```

### E6: Network Drop During Call

```
Trigger: WiFi disconnects, network unreachable during active call
Detection: WebSocket onFailure or onClosed callback
Severity: Critical

User Experience:
┌─────────────────────────────────┐
│  Connected Screen               │
│                                 │
│   Smile Dental Receptionist     │
│   02:34                         │
│                                 │
│   ⚠️  "Connection lost"         │  ← Brief toast
│                                 │
│   → Auto-transitions to ENDED   │
│                                 │
└─────────────────────────────────┘

Recovery:
1. Stop audio recording and playback
2. Transition to call ended screen
3. Show duration of call before disconnect
4. User can redial when network restored
```

### E7: Microphone Permission Denied

```
Trigger: User denies RECORD_AUDIO permission (API 23+)
Detection: onRequestPermissionsResult with PERMISSION_DENIED
Severity: Critical

User Experience:
┌─────────────────────────────────┐
│  Dialer Screen                  │
│                                 │
│  ⚠️  "Microphone permission     │  ← Toast or dialog
│      denied. Calls require      │
│      microphone access."        │
│                                 │
│  [OK] → Opens App Settings      │
│                                 │
│  [Keypad...]                    │
└─────────────────────────────────┘

Recovery:
1. Stay on dialer screen
2. Offer to open app settings
3. User grants permission in Settings
4. User returns to app and retries
```

### E8: AudioRecord Initialization Failure

```
Trigger: Audio hardware unavailable, another app using mic
Detection: audioRecord.getState() != AudioRecord.STATE_INITIALIZED
Severity: Critical

User Experience:
┌─────────────────────────────────┐
│  Call Ended Screen              │
│                                 │
│       ❌                        │
│   Call Ended                    │
│                                 │
│   Audio system unavailable      │
│   Close other apps using        │
│   the microphone                │
│                                 │
│   Duration: 0s                  │
│                                 │
│   [ Back to Dialer ]            │
└─────────────────────────────────┘

Recovery:
1. Stop call attempt
2. Show specific error message
3. User closes competing app
4. User retries call
```

---

## 11.4 Error Scenarios — Backend Server

### E9: Invalid JSON Message

```
Trigger: Client sends malformed JSON
Detection: JSON.parse() throws exception
Severity: Minor

Server Response:
{ "event": "error", "reason": "Invalid JSON" }

Recovery:
1. Log error for debugging
2. Send error response to client
3. Continue processing other messages
```

### E10: Agent Disconnects During Call

```
Trigger: Agent WebSocket closes unexpectedly during active call
Detection: ws.on('close') on agent connection
Severity: Critical

Server Actions:
1. Remove agent from agents map
2. Find active call for this agent
3. Send {event: "ended", reason: "agent_disconnect"} to mobile
4. Clean up call record
5. Log: "Agent disconnected during call"

Mobile Experience:
- Receives "ended" event
- Transitions to call ended screen
- Shows "Agent disconnected"
```

### E11: Mobile Disconnects During Call

```
Trigger: Mobile app crashes or user closes app during call
Detection: ws.on('close') on mobile connection
Severity: Major

Server Actions:
1. Find active call for this mobile connection
2. Send {event: "ended", reason: "client_disconnect"} to agent
3. Clean up call record
4. Log: "Mobile disconnected during call"

Agent Experience:
- Receives "ended" event
- Stops audio processing
- Returns to idle state
```

### E12: Duplicate Agent Registration

```
Trigger: Two agents try to register with the same number
Detection: agents.has(number) returns true
Severity: Minor

Server Response:
{ "event": "error", "reason": "Number already registered" }

Recovery:
1. Reject second registration
2. First agent remains active
3. Second agent must use different number
```

---

## 11.5 Error Recovery Patterns

### Retry Strategy

| Error | Retry? | Strategy |
|-------|--------|----------|
| Connection timeout | Yes | User-initiated retry |
| Auth failure | No | Requires config change |
| Number not found | Yes | User can retry with different number |
| Agent unavailable | Yes | Agent may reconnect |
| Network drop | Yes | User redials after network restore |
| Permission denied | No | Requires user action in Settings |
| Audio init failure | Yes | After closing competing app |

### Graceful Degradation

```
┌─────────────────────────────────────────────────────────────┐
│              Graceful Degradation Hierarchy                  │
│                                                             │
│  Full Functionality                                         │
│  ├── Audio streaming + control messages                     │
│  ├── Waveform visualization                                 │
│  ├── Call timer                                             │
│  └── Mute/speaker controls                                  │
│                                                             │
│  Degraded: No Waveform (API 16-18)                          │
│  ├── Audio streaming works                                  │
│  ├── Simulated waveform animation                           │
│  └── All controls work                                      │
│                                                             │
│  Degraded: No Speaker Toggle                                │
│  ├── Audio streaming works                                  │
│  └── Uses default audio routing                             │
│                                                             │
│  Degraded: Network Unstable                                 │
│  ├── Audio may have gaps                                    │
│  ├── WebSocket reconnects automatically                     │
│  └── Call may drop if reconnect fails                       │
│                                                             │
│  Minimal: Server Unreachable                                │
│  ├── App still launches                                     │
│  ├── Dialer UI works                                        │
│  └── Calls fail with clear error message                    │
│                                                             │
│  Critical Failure: Audio Hardware Unavailable               │
│  ├── App launches                                           │
│  └── Calls fail with specific error                         │
└─────────────────────────────────────────────────────────────┘
```

---

## 11.6 Logging Strategy

### Android App Logging

```java
// Use Android Log for debugging
private static final String TAG = "SmileDental";

// Info: Normal operations
Log.i(TAG, "Call initiated: number=" + number);
Log.i(TAG, "Connected to agent: " + agentName);
Log.i(TAG, "Call ended: duration=" + duration + "s");

// Warning: Recoverable issues
Log.w(TAG, "Server unreachable, retrying...");
Log.w(TAG, "Audio buffer underrun detected");

// Error: Failures
Log.e(TAG, "AudioRecord initialization failed", exception);
Log.e(TAG, "WebSocket connection error: " + message, exception);
```

### Backend Logging

```javascript
// Console logging with timestamps
function log(level, message, data = null) {
    const timestamp = new Date().toISOString();
    const prefix = `[${timestamp}] [${level.toUpperCase()}]`;

    if (data) {
        console[level === 'error' ? 'error' : 'log'](
            `${prefix} ${message}`, JSON.stringify(data)
        );
    } else {
        console[level === 'error' ? 'error' : 'log'](
            `${prefix} ${message}`
        );
    }
}

// Usage
log('info', 'Agent registered', { number: '101', name: 'Receptionist' });
log('info', 'Call started', { callId: 1, number: '101' });
log('warn', 'Agent disconnected during call', { callId: 1 });
log('error', 'Failed to parse message', { raw: data.toString() });
```

---

*Next: [Page 12 — Future Roadmap](12-future-roadmap.md)*
