# Page 04 — WebSocket Protocol

---

## 4.1 Connection Endpoints

| Endpoint | Purpose | Client Type |
|----------|---------|-------------|
| `ws://<host>:3000/call` | Mobile phone client | Android app |
| `ws://<host>:3000/agent` | AI agent client | Node.js agent services |

Both endpoints use the same WebSocket server but handle different message types and behaviors.

---

## 4.2 Message Format

All control messages are **JSON text frames**. Audio data is sent as **binary frames** (raw PCM).

### JSON Message Structure
```json
{
    "action": "dial",        // or "hangup", "mute", "register", etc.
    "number": "101",         // phone number (for dial)
    "token": "user-token",   // auth token (for dial)
    "muted": true            // mute state (for mute toggle)
}
```

### Binary Frame Structure
```
┌─────────────────────────────────────────────┐
│  WebSocket Binary Frame (opcode = 0x2)      │
│                                             │
│  Payload: Raw PCM 16-bit mono 16kHz audio   │
│  Typical size: 320-640 bytes (10-20ms)      │
│  No header, no metadata — pure PCM data     │
└─────────────────────────────────────────────┘
```

---

## 4.3 Mobile Client Protocol (ws://host:3000/call)

### Client → Server Messages

#### 1. Dial
Sent immediately after WebSocket connection opens.
```json
{
    "action": "dial",
    "number": "101"
}
```

#### 2. Hang Up
Sent when user presses hang-up button or cancels call.
```json
{
    "action": "hangup"
}
```

#### 3. Mute Toggle
Sent when user toggles mute.
```json
{
    "action": "mute",
    "muted": true
}
```

### Server → Client Messages

#### 1. Ringing
Backend found the agent and is ringing.
```json
{
    "event": "ringing",
    "agentName": "Smile Dental Receptionist"
}
```

#### 2. Connected
Agent accepted the call. Begin audio streaming.
```json
{
    "event": "connected"
}
```

#### 3. Busy / Not Found
Number doesn't exist or agent is unavailable.
```json
{
    "event": "busy",
    "reason": "Number not available on this network"
}
```

#### 4. Ended
Call has ended (agent hung up, timeout, error).
```json
{
    "event": "ended",
    "reason": "agent_hangup"
}
```

Possible reasons:
| Reason | Description |
|--------|-------------|
| `agent_hangup` | Agent disconnected |
| `timeout` | No answer within time limit |
| `error` | Server-side error |
| `client_timeout` | Mobile client stopped sending audio |

---

## 4.4 Agent Client Protocol (ws://host:3000/agent)

### Client → Server Messages

#### 1. Register
Sent when agent connects to establish its phone number.
```json
{
    "action": "register",
    "number": "101",
    "name": "Smile Dental Receptionist"
}
```

#### 2. Accept Call
Sent when agent accepts an incoming call.
```json
{
    "action": "accept"
}
```

#### 3. Reject Call
Sent when agent rejects an incoming call.
```json
{
    "action": "reject",
    "reason": "busy"
}
```

#### 4. Hang Up
Sent when agent ends an active call.
```json
{
    "action": "hangup"
}
```

### Server → Agent Messages

#### 1. Registration Confirmed
```json
{
    "event": "registered",
    "number": "101",
    "name": "Smile Dental Receptionist"
}
```

#### 2. Incoming Call
```json
{
    "event": "incoming_call",
    "callerId": "user-token-123"
}
```

#### 3. Call Accepted (by other side)
```json
{
    "event": "call_accepted"
}
```

#### 4. Call Ended
```json
{
    "event": "ended",
    "reason": "client_hangup"
}
```

---

## 4.5 Complete Call Sequence

```
Mobile                          Backend                         Agent
  │                               │                               │
  │── WS Connect ────────────────►│                               │
  │   /call                       │                               │
  │                               │                               │
  │── {dial, "101"} ─────────────►│                               │
  │                               │                               │
  │                               │── Lookup "101" ──────────────  │
  │                               │── Found: Agent WS             │
  │                               │                               │
  │                               │── {incoming_call} ───────────►│
  │                               │                               │
  │                               │◄── {accept} ──────────────────│
  │                               │                               │
  │◄── {ringing, agentName} ─────│                               │
  │                               │                               │
  │◄── {connected} ──────────────│                               │
  │                               │                               │
  │── PCM binary ────────────────►│── PCM binary ────────────────►│
  │◄── PCM binary ───────────────│◄── PCM binary ────────────────│
  │                               │                               │
  │   (audio streaming both ways) │                               │
  │                               │                               │
  │── {hangup} ──────────────────►│                               │
  │                               │── {ended} ───────────────────►│
  │                               │── Cleanup ─────────────────── │
  │◄── WS Close ─────────────────│                               │
  │                               │                               │
```

---

## 4.6 WebSocketManager Implementation

```java
public class WebSocketManager {

    private static final String TAG = "WebSocketManager";

    private OkHttpClient client;
    private WebSocket webSocket;
    private CallListener listener;

    private boolean isConnected = false;

    public interface CallListener {
        void onRinging(String agentName);
        void onConnected();
        void onBusy(String reason);
        void onEnded(String reason);
        void onAudioReceived(byte[] data, int length);
        void onError(String message);
    }

    public WebSocketManager(CallListener listener) {
        this.listener = listener;
        this.client = new OkHttpClient.Builder()
            .readTimeout(0, TimeUnit.MILLISECONDS)  // No timeout for WS
            .pingInterval(30, TimeUnit.SECONDS)     // Keep-alive
            .build();
    }

    public void connect(String url, String number) {
        Request request = new Request.Builder()
            .url(url)
            .build();

        webSocket = client.newWebSocket(request, new WebSocketListener() {
            @Override
            public void onOpen(WebSocket ws, Response response) {
                isConnected = true;
                // Send dial message immediately
                sendDial(number);
            }

            @Override
            public void onMessage(WebSocket ws, String text) {
                try {
                    JSONObject json = new JSONObject(text);
                    String event = json.optString("event", "");

                    switch (event) {
                        case "ringing":
                            String agentName = json.optString("agentName", "");
                            listener.onRinging(agentName);
                            break;
                        case "connected":
                            listener.onConnected();
                            break;
                        case "busy":
                            String reason = json.optString("reason", "");
                            listener.onBusy(reason);
                            break;
                        case "ended":
                            String endReason = json.optString("reason", "");
                            listener.onEnded(endReason);
                            break;
                    }
                } catch (JSONException e) {
                    Log.e(TAG, "Failed to parse message: " + text, e);
                }
            }

            @Override
            public void onMessage(WebSocket ws, ByteString bytes) {
                // Binary audio data
                byte[] data = bytes.toByteArray();
                listener.onAudioReceived(data, data.length);
            }

            @Override
            public void onFailure(WebSocket ws, Throwable t, Response response) {
                isConnected = false;
                String msg = t != null ? t.getMessage() : "Connection failed";
                listener.onError(msg);
            }

            @Override
            public void onClosed(WebSocket ws, int code, String reason) {
                isConnected = false;
                listener.onEnded("connection_closed");
            }
        });
    }

    private void sendDial(String number) {
        try {
            JSONObject json = new JSONObject();
            json.put("action", "dial");
            json.put("number", number);
            webSocket.send(json.toString());
        } catch (JSONException e) {
            Log.e(TAG, "Failed to create dial message", e);
        }
    }

    public void sendAudio(byte[] data, int length) {
        if (webSocket != null && isConnected) {
            webSocket.send(ByteString.of(data, 0, length));
        }
    }

    public void sendHangup() {
        if (webSocket != null && isConnected) {
            try {
                JSONObject json = new JSONObject();
                json.put("action", "hangup");
                webSocket.send(json.toString());
            } catch (JSONException e) {
                Log.e(TAG, "Failed to create hangup message", e);
            }
        }
    }

    public void sendMute(boolean muted) {
        if (webSocket != null && isConnected) {
            try {
                JSONObject json = new JSONObject();
                json.put("action", "mute");
                json.put("muted", muted);
                webSocket.send(json.toString());
            } catch (JSONException e) {
                Log.e(TAG, "Failed to create mute message", e);
            }
        }
    }

    public void disconnect() {
        if (webSocket != null) {
            webSocket.close(1000, "Normal closure");
            webSocket = null;
        }
        isConnected = false;
    }
}
```

---

## 4.7 Error Handling

| Error | Trigger | Client Action |
|-------|---------|---------------|
| Connection refused | Server not running | Show toast, return to dialer |
| Connection timeout | Network unreachable | Show "Connection timeout", return to dialer |
| Invalid token | Auth failure | Show "Authentication failed", return to dialer |
| Number not found | Phonebook miss | Show "Number not available" |
| Agent busy | Agent in another call | Show "Agent is busy", return to dialer |
| Unexpected close | Network drop | Show "Connection lost", transition to ENDED |
| Parse error | Malformed JSON | Log error, ignore message |

---

*Next: [Page 05 — Backend Server Design](05-backend-server-design.md)*
