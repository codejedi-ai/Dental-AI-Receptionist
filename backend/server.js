const WebSocket = require('ws');
const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');

require('dotenv').config();

const PORT = parseInt(process.env.PORT || '3000', 10);

// Load phonebook
const phonebook = JSON.parse(
  fs.readFileSync(path.join(__dirname, 'phonebook.json'), 'utf8')
);

// In-memory state
const agents = new Map();      // number -> WebSocket
const activeCalls = new Map(); // callId -> { mobileWs, agentWs, number, agentName, startTime }
let callIdCounter = 0;

const app = express();
const server = http.createServer(app);
const wss = new WebSocket.Server({ noServer: true });

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    agents: Array.from(agents.keys()),
    activeCalls: activeCalls.size,
    phonebook
  });
});

// Route WebSocket upgrades by path
server.on('upgrade', (request, socket, head) => {
  const { pathname } = new URL(request.url, `http://localhost:${PORT}`);

  if (pathname === '/call') {
    wss.handleUpgrade(request, socket, head, (ws) => handleMobileClient(ws));
  } else if (pathname === '/agent') {
    wss.handleUpgrade(request, socket, head, (ws) => handleAgentClient(ws));
  } else {
    socket.destroy();
  }
});

// ─────────────────────────────────────────────
// Mobile Client Handler  (/call)
// ─────────────────────────────────────────────
function handleMobileClient(ws) {
  let agentWs = null;
  let dialedNumber = null;

  ws.on('message', (data) => {
    // Binary = PCM audio — forward to agent
    if (Buffer.isBuffer(data)) {
      if (agentWs && agentWs.readyState === WebSocket.OPEN) {
        agentWs.send(data, { binary: true });
      }
      return;
    }

    let msg;
    try { msg = JSON.parse(data.toString()); }
    catch { ws.send(JSON.stringify({ event: 'error', reason: 'Invalid JSON' })); return; }

    switch (msg.action) {
      case 'dial':
        handleDial(ws, msg, (foundAgent) => {
          agentWs = foundAgent;
          dialedNumber = msg.number;
        });
        break;

      case 'hangup':
        handleMobileHangup(ws);
        break;

      case 'mute':
        if (agentWs && agentWs.readyState === WebSocket.OPEN) {
          agentWs.send(JSON.stringify({ event: 'mute', muted: msg.muted }));
        }
        break;

      default:
        ws.send(JSON.stringify({ event: 'error', reason: 'Unknown action' }));
    }
  });

  ws.on('close', () => handleMobileDisconnect(ws));
  ws.on('error', (err) => console.error('Mobile WS error:', err.message));
}

function handleDial(mobileWs, msg, setAgent) {
  const { number } = msg;

  const agentName = phonebook[number];
  if (!agentName) {
    mobileWs.send(JSON.stringify({ event: 'busy', reason: 'Number not available on this network' }));
    mobileWs.close(1000, 'Number not found');
    return;
  }

  const agentWs = agents.get(number);
  if (!agentWs || agentWs.readyState !== WebSocket.OPEN) {
    mobileWs.send(JSON.stringify({ event: 'busy', reason: `${agentName} is not currently available` }));
    mobileWs.close(1000, 'Agent unavailable');
    return;
  }

  const callId = ++callIdCounter;
  activeCalls.set(callId, { mobileWs, agentWs, number, agentName, startTime: Date.now() });

  // Notify agent
  agentWs.send(JSON.stringify({ event: 'incoming_call', callId }));

  // Ring the mobile
  mobileWs.send(JSON.stringify({ event: 'ringing', agentName }));

  // Connect after 1s (gives agent time to accept)
  setTimeout(() => {
    if (mobileWs.readyState === WebSocket.OPEN && agentWs.readyState === WebSocket.OPEN) {
      mobileWs.send(JSON.stringify({ event: 'connected' }));
      agentWs.send(JSON.stringify({ event: 'call_accepted', callId }));
      console.log(`[Call ${callId}] Connected: ${number} (${agentName})`);
    }
  }, 1000);

  setAgent(agentWs);
}

function handleMobileHangup(mobileWs) {
  for (const [callId, call] of activeCalls) {
    if (call.mobileWs === mobileWs) {
      if (call.agentWs.readyState === WebSocket.OPEN) {
        call.agentWs.send(JSON.stringify({ event: 'ended', reason: 'client_hangup' }));
      }
      activeCalls.delete(callId);
      console.log(`[Call ${callId}] Ended: mobile hangup`);
      break;
    }
  }
  mobileWs.close(1000, 'Call ended');
}

function handleMobileDisconnect(mobileWs) {
  for (const [callId, call] of activeCalls) {
    if (call.mobileWs === mobileWs) {
      if (call.agentWs.readyState === WebSocket.OPEN) {
        call.agentWs.send(JSON.stringify({ event: 'ended', reason: 'client_disconnect' }));
      }
      activeCalls.delete(callId);
      console.log(`[Call ${callId}] Ended: mobile disconnected`);
      break;
    }
  }
}

// ─────────────────────────────────────────────
// Agent Client Handler  (/agent)
// ─────────────────────────────────────────────
function handleAgentClient(ws) {
  let registeredNumber = null;

  ws.on('message', (data) => {
    // Binary = PCM audio — forward to mobile
    if (Buffer.isBuffer(data)) {
      for (const [, call] of activeCalls) {
        if (call.agentWs === ws && call.mobileWs.readyState === WebSocket.OPEN) {
          call.mobileWs.send(data, { binary: true });
          return;
        }
      }
      return;
    }

    let msg;
    try { msg = JSON.parse(data.toString()); }
    catch { ws.send(JSON.stringify({ event: 'error', reason: 'Invalid JSON' })); return; }

    switch (msg.action) {
      case 'register':
        registeredNumber = msg.number;
        agents.set(msg.number, ws);
        ws.send(JSON.stringify({ event: 'registered', number: msg.number, name: msg.name }));
        console.log(`[Agent] Registered: ${msg.number} (${msg.name})`);
        break;

      case 'accept':
        // Agent explicitly accepted — already handled by auto-connect timeout above
        break;

      case 'reject':
        for (const [callId, call] of activeCalls) {
          if (call.agentWs === ws) {
            call.mobileWs.send(JSON.stringify({ event: 'busy', reason: msg.reason || 'Agent rejected call' }));
            activeCalls.delete(callId);
            break;
          }
        }
        break;

      case 'hangup':
        for (const [callId, call] of activeCalls) {
          if (call.agentWs === ws) {
            call.mobileWs.send(JSON.stringify({ event: 'ended', reason: 'agent_hangup' }));
            activeCalls.delete(callId);
            console.log(`[Call ${callId}] Ended: agent hangup`);
            break;
          }
        }
        break;
    }
  });

  ws.on('close', () => {
    if (registeredNumber) {
      agents.delete(registeredNumber);
      console.log(`[Agent] Disconnected: ${registeredNumber}`);

      for (const [callId, call] of activeCalls) {
        if (call.agentWs === ws) {
          call.mobileWs.send(JSON.stringify({ event: 'ended', reason: 'agent_disconnect' }));
          activeCalls.delete(callId);
          break;
        }
      }
    }
  });

  ws.on('error', (err) => console.error('Agent WS error:', err.message));
}

// ─────────────────────────────────────────────
// Start
// ─────────────────────────────────────────────
server.listen(PORT, '0.0.0.0', () => {
  console.log('🦷 Smile Dental Backend Server');
  console.log(`   WebSocket: ws://0.0.0.0:${PORT}`);
  console.log(`     /call  → mobile clients (iOS app)`);
  console.log(`     /agent → AI agent services`);
  console.log(`   Health: http://0.0.0.0:${PORT}/health`);
  console.log(`   Phonebook: ${Object.keys(phonebook).join(', ')}`);
});

process.on('SIGINT', () => {
  console.log('\nShutting down...');
  wss.clients.forEach((ws) => ws.close());
  server.close(() => process.exit(0));
});
