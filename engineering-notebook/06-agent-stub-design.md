# Page 06 — Agent Stub Design

---

## 6.1 Purpose

The Agent Stub is a **test agent** that connects to the backend, registers a phone number, and echoes received audio back to the caller (loopback). This allows testing the **full audio pipeline** — from Android microphone → WebSocket → backend → agent → backend → Android speaker — without needing a real AI service.

### Use Cases
1. **Audio pipeline verification** — Confirm PCM audio flows end-to-end
2. **Latency measurement** — Echo round-trip time = network latency × 2
3. **UI testing** — Test all call states without AI complexity
4. **Stress testing** — Verify stability under extended call duration

---

## 6.2 File Structure

```
agent-stub/
├── agent.js            # Main agent: connects, registers, echoes
├── package.json        # Dependencies
└── README.md           # Usage instructions
```

---

## 6.3 agent.js — Full Implementation

```javascript
const WebSocket = require('ws');

// Configuration
const BACKEND_URL = process.env.BACKEND_URL || 'ws://localhost:3000/agent';
const AGENT_NUMBER = process.env.AGENT_NUMBER || '101';
const AGENT_NAME = process.env.AGENT_NAME || 'Smile Dental Receptionist';

// State
let ws = null;
let currentCall = null;
let audioStats = {
    framesReceived: 0,
    bytesReceived: 0,
    framesSent: 0,
    bytesSent: 0,
    startTime: null
};

// ──────────────────────────────────────────────
// Connect to Backend
// ──────────────────────────────────────────────
function connect() {
    console.log(`🤖 Agent connecting to ${BACKEND_URL}`);
    console.log(`   Number: ${AGENT_NUMBER}`);
    console.log(`   Name: ${AGENT_NAME}`);

    ws = new WebSocket(BACKEND_URL);

    ws.on('open', () => {
        console.log('✅ Connected to backend');
        register();
    });

    ws.on('message', (data) => {
        // Binary = audio data from mobile
        if (Buffer.isBuffer(data)) {
            handleAudio(data);
            return;
        }

        // Text = JSON control message
        try {
            const msg = JSON.parse(data.toString());
            handleMessage(msg);
        } catch (e) {
            console.error('Failed to parse message:', data.toString());
        }
    });

    ws.on('close', (code, reason) => {
        console.log(`❌ Disconnected: code=${code}, reason=${reason}`);
        currentCall = null;
        // Attempt reconnection after 3 seconds
        setTimeout(connect, 3000);
    });

    ws.on('error', (err) => {
        console.error('WebSocket error:', err.message);
    });
}

// ──────────────────────────────────────────────
// Register Agent
// ──────────────────────────────────────────────
function register() {
    const msg = {
        action: 'register',
        number: AGENT_NUMBER,
        name: AGENT_NAME
    };
    ws.send(JSON.stringify(msg));
}

// ──────────────────────────────────────────────
// Handle Control Messages
// ──────────────────────────────────────────────
function handleMessage(msg) {
    switch (msg.event) {
        case 'registered':
            console.log(`✅ Registered as ${msg.number} (${msg.name})`);
            console.log('   Waiting for incoming calls...');
            break;

        case 'incoming_call':
            console.log(`📞 Incoming call from ${msg.callerId} (callId: ${msg.callId})`);
            currentCall = { callId: msg.callId, callerId: msg.callerId };
            audioStats = {
                framesReceived: 0,
                bytesReceived: 0,
                framesSent: 0,
                bytesSent: 0,
                startTime: Date.now()
            };
            acceptCall();
            break;

        case 'call_accepted':
            console.log('🔗 Call accepted — audio bridge active');
            break;

        case 'ended':
            console.log(`📴 Call ended: ${msg.reason}`);
            if (audioStats.startTime) {
                const duration = ((Date.now() - audioStats.startTime) / 1000).toFixed(1);
                console.log(`   Duration: ${duration}s`);
                console.log(`   Audio stats:`);
                console.log(`     Received: ${audioStats.framesReceived} frames (${audioStats.bytesReceived} bytes)`);
                console.log(`     Sent: ${audioStats.framesSent} frames (${audioStats.bytesSent} bytes)`);
            }
            currentCall = null;
            break;

        case 'mute':
            console.log(`🔇 Mute state: ${msg.muted ? 'ON' : 'OFF'}`);
            break;

        case 'error':
            console.error(`⚠️  Error from backend: ${msg.reason}`);
            break;

        default:
            console.log('Unknown message:', msg);
    }
}

// ──────────────────────────────────────────────
// Accept Call
// ──────────────────────────────────────────────
function acceptCall() {
    const msg = {
        action: 'accept'
    };
    ws.send(JSON.stringify(msg));
    console.log('   Call accepted');
}

// ──────────────────────────────────────────────
// Handle Audio (Loopback Echo)
// ──────────────────────────────────────────────
function handleAudio(data) {
    audioStats.framesReceived++;
    audioStats.bytesReceived += data.length;

    // Echo the audio back immediately (loopback)
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(data, { binary: true });
        audioStats.framesSent++;
        audioStats.bytesSent += data.length;
    }
}

// ──────────────────────────────────────────────
// Graceful Shutdown
// ──────────────────────────────────────────────
function shutdown() {
    console.log('\n🛑 Shutting down agent...');
    if (ws) {
        ws.close(1000, 'Agent shutdown');
    }
    process.exit(0);
}

process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

// ──────────────────────────────────────────────
// Start
// ──────────────────────────────────────────────
connect();
```

---

## 6.4 package.json

```json
{
    "name": "smile-dental-agent-stub",
    "version": "1.0.0",
    "description": "Test agent for Smile Dental AI Phone Dialer — audio loopback",
    "main": "agent.js",
    "scripts": {
        "start": "node agent.js",
        "start:101": "AGENT_NUMBER=101 AGENT_NAME='Smile Dental Receptionist' node agent.js",
        "start:102": "AGENT_NUMBER=102 AGENT_NAME='Billing Agent' node agent.js",
        "start:103": "AGENT_NUMBER=103 AGENT_NAME='Appointment Reminder' node agent.js"
    },
    "dependencies": {
        "ws": "^8.14.0"
    },
    "engines": {
        "node": ">=14.0.0"
    }
}
```

---

## 6.5 Agent Behavior Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Agent Lifecycle                            │
│                                                             │
│  ┌──────────┐    ┌───────────┐    ┌──────────────┐         │
│  │ Connect  │───►│ Register  │───►│  Idle/Wait   │         │
│  │ to WS    │    │ number    │    │  for calls   │         │
│  └──────────┘    └───────────┘    └──────┬───────┘         │
│                                          │                  │
│                              incoming_call│                  │
│                                          │                  │
│                                   ┌──────▼───────┐         │
│                                   │   Accept     │         │
│                                   │   Call       │         │
│                                   └──────┬───────┘         │
│                                          │                  │
│                                    ┌─────▼──────┐          │
│                                    │  Audio     │          │
│                                    │  Loopback  │          │
│                                    │  (echo)    │          │
│                                    └─────┬──────┘          │
│                                          │                  │
│                                    ended │                  │
│                                          │                  │
│                                   ┌──────▼───────┐         │
│                                   │  Log Stats   │         │
│                                   │  Reset       │         │
│                                   └──────┬───────┘         │
│                                          │                  │
│                                          ▼                  │
│                                   ┌──────────────┐         │
│                                   │  Idle/Wait   │         │
│                                   │  for calls   │         │
│                                   └──────────────┘         │
│                                                             │
│  On disconnect: auto-reconnect after 3 seconds              │
└─────────────────────────────────────────────────────────────┘
```

---

## 6.6 Running Multiple Agents

```bash
# Terminal 1: Backend
cd backend && npm install && node server.js

# Terminal 2: Agent 101 (Receptionist)
cd agent-stub && npm install
AGENT_NUMBER=101 AGENT_NAME="Smile Dental Receptionist" node agent.js

# Terminal 3: Agent 102 (Billing)
cd agent-stub && npm install
AGENT_NUMBER=102 AGENT_NAME="Billing Agent" node agent.js

# Terminal 4: Agent 103 (Appointments)
cd agent-stub && npm install
AGENT_NUMBER=103 AGENT_NAME="Appointment Reminder Agent" node agent.js
```

Each agent runs as a **separate process** connecting to the same backend. The backend maintains a map of all registered agents.

---

## 6.7 Audio Loopback Verification

When testing, verify:

| Check | Method | Expected |
|-------|--------|----------|
| Audio capture | Speak into phone mic | Agent receives PCM data |
| Audio echo | Agent sends back | Phone plays audio through speaker |
| Round-trip latency | Measure delay | < 200ms on local WiFi |
| No audio distortion | Listen to echo | Clear, no clipping |
| Mute works | Toggle mute | Echo stops during mute |
| Hangup works | Press hangup | Audio stops, call ends |

---

## 6.8 Future: Real Agent Integration

To replace the stub with a real AI agent:

1. **Speech-to-Text:** Pipe received PCM audio to STT service (Whisper, Google STT)
2. **LLM Processing:** Send transcribed text to AI model for response generation
3. **Text-to-Speech:** Convert AI response to PCM audio
4. **Send PCM back:** Same as loopback, but with generated audio instead of echo

```javascript
// Pseudocode for real agent
async function handleAudio(data) {
    // 1. Accumulate audio buffer
    audioBuffer.push(data);

    // 2. Detect end of speech (VAD)
    if (isEndOfSpeech(audioBuffer)) {
        // 3. STT
        const text = await speechToText(audioBuffer);

        // 4. LLM response
        const response = await llm.generate(text);

        // 5. TTS
        const responseAudio = await textToSpeech(response);

        // 6. Send back
        ws.send(responseAudio, { binary: true });

        // 7. Reset buffer
        audioBuffer = [];
    }
}
```

---

*Next: [Page 07 — API Compatibility](07-api-compatibility.md)*
