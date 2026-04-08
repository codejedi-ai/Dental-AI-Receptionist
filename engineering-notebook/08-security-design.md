# Page 08 — Security Design

---

## 8.1 Overview

**Current state: No authentication.** The system operates as an open service on the local network. Anyone who can reach the backend server's IP and port can make calls.

This is acceptable for the **frontend-only prototype phase** where the system runs on an isolated office WiFi network.

---

## 8.2 Threat Model

```
┌─────────────────────────────────────────────────────────────┐
│                    Trust Boundaries                         │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              TRUSTED: Local Network                 │   │
│  │                                                     │   │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐      │   │
│  │  │  Phone   │    │ Backend  │    │  Agents  │      │   │
│  │  │  (App)   │◄──►│  Server  │◄──►│  (AI)    │      │   │
│  │  └──────────┘    └──────────┘    └──────────┘      │   │
│  │                                                     │   │
│  │  All communication: unencrypted WebSocket (ws://)   │   │
│  │  No authentication — open access on local network   │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Protection relies on:                                      │
│  • WiFi network isolation (WPA2 password)                   │
│  • No external port forwarding                              │
│  • Physical access control to the network                   │
└─────────────────────────────────────────────────────────────┘
```

### Current Threats
| Threat | Likelihood | Impact | Mitigation |
|--------|-----------|--------|------------|
| Unauthorized dial access | Low (local net) | Low | WiFi password |
| Eavesdropping on audio | Low (local net) | Medium | Network isolation |
| DoS (flood calls) | Low | Medium | Rate limiting (future) |

---

## 8.3 Android Permissions

### Required Permissions

| Permission | Purpose | Protection Level |
|-----------|---------|-----------------|
| `INTERNET` | WebSocket communication | Normal (auto-granted) |
| `RECORD_AUDIO` | Microphone capture | Dangerous (runtime on API 23+) |
| `MODIFY_AUDIO_SETTINGS` | Speakerphone toggle | Normal (auto-granted) |
| `ACCESS_NETWORK_STATE` | Network connectivity check | Normal (auto-granted) |

### Permission Declaration
```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.RECORD_AUDIO" />
<uses-permission android:name="android.permission.MODIFY_AUDIO_SETTINGS" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

### Runtime Permission Handling (API 23+)

On Android 6.0+ (API 23+), `RECORD_AUDIO` requires runtime request:

```java
private static final int PERMISSION_REQUEST_MIC = 1001;

private boolean ensureMicPermission() {
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        if (checkSelfPermission(Manifest.permission.RECORD_AUDIO)
                != PackageManager.PERMISSION_GRANTED) {
            requestPermissions(
                new String[]{Manifest.permission.RECORD_AUDIO},
                PERMISSION_REQUEST_MIC
            );
            return false;
        }
    }
    return true;
}
```

On Android 4.1-5.1 (API 16-22), all permissions are granted at install time — no runtime request needed.

---

## 8.4 Network Security

### Current: Unencrypted (ws://)

All WebSocket communication is **unencrypted**. Audio and control messages travel as plain text on the local network.

```
ws://192.168.1.100:3000/call
     ↑
     Unencrypted — acceptable for isolated local network
```

### Why This Is Acceptable (For Now)

1. **Local network only** — traffic doesn't leave the premises
2. **WiFi isolation** — external attackers can't access the network
3. **Low-value target** — dental appointment audio isn't high-value
4. **Performance** — no encryption overhead on low-end devices

### Future: Encrypted (wss://)

For production or wider deployment, upgrade to `wss://` with TLS.

---

## 8.5 Input Validation

### Backend Input Sanitization

```javascript
// Validate phone number format (digits, *, # — 1 to 10 chars)
function isValidPhoneNumber(number) {
    return /^[0-9*#]{1,10}$/.test(number);
}

// In dial handler
if (!isValidPhoneNumber(number)) {
    mobileWs.send(JSON.stringify({
        event: 'busy',
        reason: 'Invalid phone number'
    }));
    return;
}
```

---

## 8.6 Future: Adding Authentication

When ready to add auth, the recommended approach is:

1. **Token-based auth** — simple shared-secret tokens in `.env`
2. **Token sent with dial message** — `{ "action": "dial", "number": "101", "token": "..." }`
3. **Backend validates token** — rejects unknown tokens
4. **TLS (wss://)** — encrypt all traffic

This can be layered on top of the existing system without changing the audio pipeline.

---

*Next: [Page 09 — Build & Deployment](09-build-deployment.md)*
