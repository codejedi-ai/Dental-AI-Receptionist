# Page 05 — Backend Server Design

---

## 5.1 Overview

The backend server is a **single-process Node.js application** that handles:
1. WebSocket connections from mobile clients and AI agents
2. Authentication via token validation
3. Phone number lookup and call routing
4. Real-time bidirectional audio bridging
5. Agent registration and lifecycle management

### Tech Stack
| Component | Library | Version |
|-----------|---------|---------|
| WebSocket Server | `ws` | ^8.0.0 |
| HTTP Server (auth) | `express` | ^4.18.0 |
| Environment | `dotenv` | ^16.0.0 |
| Logging | `console` (built-in) | — |
| Node.js | — | 14+ |

### File Structure
```
backend/
├── server.js           # Main entry: WS server + Express + routing
├── phonebook.json      # Static phone number → agent name mapping
├── .env                # Environment variables (tokens, port)
├── .env.example        # Template for .env
├── package.json        # Dependencies
└── README.md           # Setup and run instructions
```

---

## 5.2 Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Backend Server (:3000)                       │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    HTTP Server (Express)                   │  │
│  │  POST /auth → { valid: true/false }                       │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                  WebSocket Server (ws)                     │  │
│  │                                                           │  │
│  │  Path: /call                    Path: /agent              │  │
│  │  ┌──────────────────┐          ┌──────────────────┐      │  │
│  │  │  Mobile Clients  │          │   Agent Clients   │      │  │
│  │  │  (Android apps)  │          │  (AI services)    │      │  │
│  │  └────────┬─────────┘          └────────┬─────────┘      │  │
│  │           │                             │                 │  │
│  │           ▼                             ▼                 │  │
│  │  ┌──────────────────┐          ┌──────────────────┐      │  │
│  │  │  Call Handler    │          │ Agent Registrar  │      │  │
│  │  │  • Auth check    │          │ • Register num   │      │  │
│  │  │  • Phone lookup  │          │ • Maintain map   │      │  │
│  │  │  • Route to agent│          │ • Heartbeat      │      │  │
│  │  └────────┬─────────┘          └──────────────────┘      │  │
│  │           │                                               │  │
│  │           ▼                                               │  │
│  │  ┌──────────────────┐                                     │  │
│  │  │   Audio Bridge   │                                     │  │
│  │  │  • Binary fwd    │                                     │  │
│  │  │  • Full-duplex   │                                     │  │
│  │  │  • No transcoding│                                     │  │
│  │  └──────────────────┘                                     │  │
│  │                                                           │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    In-Memory State                        │  │
│  │                                                           │  │
│  │  agents: Map<number, WebSocket>                           │  │
│  │  activeCalls: Map<callId, {mobile, agent, startTime}>     │  │
│  │  phonebook: { "101": "Smile Dental Receptionist" }        │  │
│  │  validTokens: ["user-token-123", "user-token-456"]        │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5.3 server.js — Full Implementation

```javascript
const WebSocket = require('ws');
const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');

// Load environment variables
require('dotenv').config();

const PORT = parseInt(process.env.PORT || '3000', 10);

// Load phonebook
const phonebookPath = path.join(__dirname, 'phonebook.json');
const phonebook = JSON.parse(fs.readFileSync(phonebookPath, 'utf8'));

// In-memory state
const agents = new Map();        // number -> WebSocket
const activeCalls = new Map();   // callId -> { mobileWs, agentWs, number, startTime }
let callIdCounter = 0;

// Create HTTP + WebSocket server
const app = express();
const server = http.createServer(app);

const wss = new WebSocket.Server({
    server: server,
    noServer: true  // We handle upgrades manually
});

// Express routes
app.use(express.json());

// Health check
app.get('/health', (req, res) => {
    res.json({
        status: 'ok',
        agents: Array.from(agents.keys()),
        activeCalls: activeCalls.size
    });
});

// Handle WebSocket upgrades
server.on('upgrade', (request, socket, head) => {
    const pathname = new URL(request.url, `http://localhost:${PORT}`).pathname;

    if (pathname === '/call') {
        wss.handleUpgrade(request, socket, head, (ws) => {
            handleMobileClient(ws);
        });
    } else if (pathname === '/agent') {
        wss.handleUpgrade(request, socket, head, (ws) => {
            handleAgentClient(ws);
        });
    } else {
        socket.destroy();
    }
});

// ──────────────────────────────────────────────
// Mobile Client Handler
// ──────────────────────────────────────────────
function handleMobileClient(ws) {
    let currentCall = null;
    let agentWs = null;
    let dialedNumber = null;

    ws.on('message', (data) => {
        // Binary data = audio
        if (Buffer.isBuffer(data)) {
            // Forward audio to agent
            if (agentWs && agentWs.readyState === WebSocket.OPEN) {
                agentWs.send(data, { binary: true });
            }
            return;
        }

        // Text data = JSON control message
        let msg;
        try {
            msg = JSON.parse(data.toString());
        } catch (e) {
            ws.send(JSON.stringify({ event: 'error', reason: 'Invalid JSON' }));
            return;
        }

        switch (msg.action) {
            case 'dial':
                handleDial(ws, msg, (foundAgentWs) => {
                    agentWs = foundAgentWs;
                    dialedNumber = msg.number;
                });
                break;

            case 'hangup':
                handleHangup(ws, dialedNumber);
                break;

            case 'mute':
                // Forward mute state to agent (optional — agent may adjust processing)
                if (agentWs && agentWs.readyState === WebSocket.OPEN) {
                    agentWs.send(JSON.stringify({ event: 'mute', muted: msg.muted }));
                }
                break;

            default:
                ws.send(JSON.stringify({ event: 'error', reason: 'Unknown action' }));
        }
    });

    ws.on('close', () => {
        handleMobileDisconnect(ws, dialedNumber);
    });

    ws.on('error', (err) => {
        console.error('Mobile WS error:', err.message);
    });
}

// ──────────────────────────────────────────────
// Handle Dial
// ──────────────────────────────────────────────
function handleDial(mobileWs, msg, setAgentWs) {
    const { number } = msg;

    // Lookup number in phonebook
    const agentName = phonebook[number];
    if (!agentName) {
        mobileWs.send(JSON.stringify({
            event: 'busy',
            reason: 'Number not available on this network'
        }));
        mobileWs.close(1000, 'Number not found');
        return;
    }

    // Find registered agent
    const agentWs = agents.get(number);
    if (!agentWs || agentWs.readyState !== WebSocket.OPEN) {
        mobileWs.send(JSON.stringify({
            event: 'busy',
            reason: `${agentName} is not currently available`
        }));
        mobileWs.close(1000, 'Agent unavailable');
        return;
    }

    // Create call record
    const callId = ++callIdCounter;
    const call = {
        id: callId,
        mobileWs: mobileWs,
        agentWs: agentWs,
        number: number,
        agentName: agentName,
        startTime: Date.now()
    };
    activeCalls.set(callId, call);

    // Notify agent of incoming call
    agentWs.send(JSON.stringify({
        event: 'incoming_call',
        callId: callId
    }));

    // Agent auto-accepts (for now — in production, wait for accept message)
    agentWs._callAccepted = true;

    // Send ringing to mobile
    mobileWs.send(JSON.stringify({
        event: 'ringing',
        agentName: agentName
    }));

    // Simulate connection after 1 second (in production, wait for agent accept)
    setTimeout(() => {
        if (mobileWs.readyState === WebSocket.OPEN &&
            agentWs.readyState === WebSocket.OPEN) {

            mobileWs.send(JSON.stringify({ event: 'connected' }));
            agentWs.send(JSON.stringify({ event: 'call_accepted' }));

            // Set up bidirectional audio bridge
            setupAudioBridge(mobileWs, agentWs, callId);
        }
    }, 1000);

    setAgentWs(agentWs);
}

// ──────────────────────────────────────────────
// Audio Bridge
// ──────────────────────────────────────────────
function setupAudioBridge(mobileWs, agentWs, callId) {
    // Audio is already bridged in the message handlers above
    // This function is a placeholder for future enhancements:
    // - Audio recording
    // - Silence detection
    // - Audio analytics
    // - Transcription
    console.log(`Audio bridge established for call ${callId}`);
}

// ──────────────────────────────────────────────
// Handle Hangup
// ──────────────────────────────────────────────
function handleHangup(mobileWs, number) {
    // Find and clean up the call
    for (const [callId, call] of activeCalls) {
        if (call.mobileWs === mobileWs) {
            // Notify agent
            if (call.agentWs && call.agentWs.readyState === WebSocket.OPEN) {
                call.agentWs.send(JSON.stringify({
                    event: 'ended',
                    reason: 'client_hangup'
                }));
            }

            // Clean up
            activeCalls.delete(callId);
            console.log(`Call ${callId} ended by mobile hangup`);
            break;
        }
    }

    mobileWs.close(1000, 'Call ended');
}

// ──────────────────────────────────────────────
// Handle Mobile Disconnect
// ──────────────────────────────────────────────
function handleMobileDisconnect(mobileWs, number) {
    for (const [callId, call] of activeCalls) {
        if (call.mobileWs === mobileWs) {
            if (call.agentWs && call.agentWs.readyState === WebSocket.OPEN) {
                call.agentWs.send(JSON.stringify({
                    event: 'ended',
                    reason: 'client_disconnect'
                }));
            }
            activeCalls.delete(callId);
            console.log(`Call ${callId} ended by mobile disconnect`);
            break;
        }
    }
}

// ──────────────────────────────────────────────
// Agent Client Handler
// ──────────────────────────────────────────────
function handleAgentClient(ws) {
    let registeredNumber = null;

    ws.on('message', (data) => {
        // Binary data = audio (forward to mobile)
        if (Buffer.isBuffer(data)) {
            for (const [callId, call] of activeCalls) {
                if (call.agentWs === ws && call.mobileWs.readyState === WebSocket.OPEN) {
                    call.mobileWs.send(data, { binary: true });
                    return;
                }
            }
            return;
        }

        // Text data = JSON control message
        let msg;
        try {
            msg = JSON.parse(data.toString());
        } catch (e) {
            ws.send(JSON.stringify({ event: 'error', reason: 'Invalid JSON' }));
            return;
        }

        switch (msg.action) {
            case 'register':
                registered_number = msg.number;
                agents.set(msg.number, ws);
                ws._agentName = msg.name;
                ws.send(JSON.stringify({
                    event: 'registered',
                    number: msg.number,
                    name: msg.name
                }));
                console.log(`Agent registered: ${msg.number} (${msg.name})`);
                break;

            case 'accept':
                ws._callAccepted = true;
                break;

            case 'reject':
                // Find associated mobile client and notify busy
                for (const [callId, call] of activeCalls) {
                    if (call.agentWs === ws) {
                        call.mobileWs.send(JSON.stringify({
                            event: 'busy',
                            reason: msg.reason || 'Agent rejected call'
                        }));
                        activeCalls.delete(callId);
                        break;
                    }
                }
                break;

            case 'hangup':
                for (const [callId, call] of activeCalls) {
                    if (call.agentWs === ws) {
                        call.mobileWs.send(JSON.stringify({
                            event: 'ended',
                            reason: 'agent_hangup'
                        }));
                        activeCalls.delete(callId);
                        break;
                    }
                }
                break;
        }
    });

    ws.on('close', () => {
        if (registered_number) {
            agents.delete(registered_number);
            console.log(`Agent disconnected: ${registered_number}`);

            // End any active calls for this agent
            for (const [callId, call] of activeCalls) {
                if (call.agentWs === ws) {
                    call.mobileWs.send(JSON.stringify({
                        event: 'ended',
                        reason: 'agent_disconnect'
                    }));
                    activeCalls.delete(callId);
                    break;
                }
            }
        }
    });

    ws.on('error', (err) => {
        console.error('Agent WS error:', err.message);
    });
}

// ──────────────────────────────────────────────
// Start Server
// ──────────────────────────────────────────────
server.listen(PORT, '0.0.0.0', () => {
    console.log(`🦷 Smile Dental Backend Server`);
    console.log(`   Listening on ws://0.0.0.0:${PORT}`);
    console.log(`   Phonebook: ${Object.keys(phonebook).length} agents`);
    console.log(`   Valid tokens: ${VALID_TOKENS.length}`);
});

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('\nShutting down...');
    wss.clients.forEach(ws => ws.close());
    server.close(() => {
        console.log('Server closed');
        process.exit(0);
    });
});
```

---

## 5.4 phonebook.json

```json
{
    "101": "Smile Dental Receptionist",
    "102": "Billing Agent",
    "103": "Appointment Reminder Agent",
    "104": "Patient Support Agent",
    "105": "Emergency Line"
}
```

---

## 5.5 .env

```env
PORT=3000
```

---

## 5.6 package.json

```json
{
    "name": "smile-dental-backend",
    "version": "1.0.0",
    "description": "WebSocket backend for Smile Dental AI Phone Dialer",
    "main": "server.js",
    "scripts": {
        "start": "node server.js",
        "dev": "node --watch server.js"
    },
    "dependencies": {
        "ws": "^8.14.0",
        "express": "^4.18.2",
        "dotenv": "^16.3.1"
    },
    "engines": {
        "node": ">=14.0.0"
    }
}
```

---

## 5.7 Call Routing Logic

```
┌─────────────────────────────────────────────────────────────┐
│                    Call Routing Flow                        │
│                                                             │
│  Mobile dials "101"                                         │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                       │
│  │ 1. Phone Lookup │  ← Is "101" in phonebook.json?       │
│  └────────┬────────┘                                       │
│           │ YES → agentName = "Smile Dental Receptionist"  │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │ 2. Agent Check  │  ← Is agent registered (agents map)?  │
│  └────────┬────────┘                                       │
│           │ YES → agentWs = agents.get("101")              │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │ 3. Notify Agent │  ← Send incoming_call to agent        │
│  └────────┬────────┘                                       │
│           │                                               │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │ 4. Ring Mobile  │  ← Send ringing + agentName           │
│  └────────┬────────┘                                       │
│           │                                               │
│           ▼                                               │
│  ┌─────────────────┐                                       │
│  │ 5. Connect      │  ← Send connected, start audio bridge │
│  └─────────────────┘                                       │
└─────────────────────────────────────────────────────────────┘
```

---

*Next: [Page 06 — Agent Stub Design](06-agent-stub-design.md)*
