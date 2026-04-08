const WebSocket = require('ws');

const BACKEND_URL = process.env.BACKEND_URL || 'ws://localhost:3000/agent';
const AGENT_NUMBER = process.env.AGENT_NUMBER || '101';
const AGENT_NAME = process.env.AGENT_NAME || 'Smile Dental Receptionist';

let ws = null;
let currentCall = null;
let audioStats = { framesReceived: 0, bytesReceived: 0, framesSent: 0, bytesSent: 0, startTime: null };

function connect() {
  console.log(`🤖 Agent connecting to ${BACKEND_URL}`);
  console.log(`   Number: ${AGENT_NUMBER}  Name: ${AGENT_NAME}`);

  ws = new WebSocket(BACKEND_URL);

  ws.on('open', () => {
    console.log('✅ Connected to backend');
    ws.send(JSON.stringify({ action: 'register', number: AGENT_NUMBER, name: AGENT_NAME }));
  });

  ws.on('message', (data) => {
    if (Buffer.isBuffer(data)) {
      audioStats.framesReceived++;
      audioStats.bytesReceived += data.length;
      // Echo audio back immediately (loopback)
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data, { binary: true });
        audioStats.framesSent++;
        audioStats.bytesSent += data.length;
      }
      return;
    }

    try {
      const msg = JSON.parse(data.toString());
      handleMessage(msg);
    } catch {
      console.error('Failed to parse message:', data.toString());
    }
  });

  ws.on('close', (code, reason) => {
    console.log(`❌ Disconnected (code=${code}). Reconnecting in 3s...`);
    currentCall = null;
    setTimeout(connect, 3000);
  });

  ws.on('error', (err) => console.error('WS error:', err.message));
}

function handleMessage(msg) {
  switch (msg.event) {
    case 'registered':
      console.log(`✅ Registered as ${msg.number} (${msg.name})`);
      console.log('   Waiting for incoming calls...');
      break;

    case 'incoming_call':
      console.log(`📞 Incoming call (callId: ${msg.callId})`);
      currentCall = { callId: msg.callId };
      audioStats = { framesReceived: 0, bytesReceived: 0, framesSent: 0, bytesSent: 0, startTime: Date.now() };
      ws.send(JSON.stringify({ action: 'accept' }));
      console.log('   Accepted');
      break;

    case 'call_accepted':
      console.log('🔗 Audio bridge active — echoing audio');
      break;

    case 'ended':
      console.log(`📴 Call ended: ${msg.reason}`);
      if (audioStats.startTime) {
        const dur = ((Date.now() - audioStats.startTime) / 1000).toFixed(1);
        console.log(`   Duration: ${dur}s | Received: ${audioStats.framesReceived} frames | Sent: ${audioStats.framesSent} frames`);
      }
      currentCall = null;
      break;

    case 'mute':
      console.log(`🔇 Mute: ${msg.muted ? 'ON' : 'OFF'}`);
      break;

    case 'error':
      console.error(`⚠️  Error: ${msg.reason}`);
      break;

    default:
      console.log('Unknown message:', msg);
  }
}

process.on('SIGINT', () => {
  console.log('\n🛑 Shutting down...');
  if (ws) ws.close(1000, 'Agent shutdown');
  process.exit(0);
});

connect();
