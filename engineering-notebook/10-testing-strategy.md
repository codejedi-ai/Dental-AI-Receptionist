# Page 10 — Testing Strategy

---

## 10.1 Overview

Testing covers three layers: **Android app**, **Backend server**, and **End-to-end integration**. The agent stub serves as the primary test double for AI agents.

---

## 10.2 Test Matrix

### Device Compatibility

| Device Type | Android Version | API | Test Status |
|------------|----------------|-----|-------------|
| Emulator (AVD) | 4.1 | 16 | ☐ Required |
| Emulator (AVD) | 4.4 | 19 | ☐ Required |
| Emulator (AVD) | 5.0 | 21 | ☐ Required |
| Emulator (AVD) | 6.0 | 23 | ☐ Required |
| Physical device | 4.1-4.3 | 16-18 | ☐ If available |
| Physical device | 10+ | 29+ | ☐ Required |

### Feature Test Matrix

| Feature | Test Case | Expected | Priority |
|---------|-----------|----------|----------|
| Dialer | Enter 101, press call | Transitions to calling screen | P0 |
| Dialer | Empty number, press call | Toast: "Enter a number first" | P0 |
| Dialer | Long-press 0 | Inserts "+" | P1 |
| Dialer | Backspace | Removes last digit | P0 |
| Dialer | Long-press backspace | Clears all digits | P1 |
| Dialer | DTMF tones | Audible tone per key press | P1 |
| Calling | Valid number → server | Shows "Calling 101..." | P0 |
| Calling | Server responds ringing | Shows "Ringing..." + agent name | P0 |
| Calling | Server responds busy | Shows error, returns to dialer | P0 |
| Calling | Press cancel | Returns to dialer, closes WS | P0 |
| Calling | Server unreachable | Timeout error, returns to dialer | P0 |
| Connected | Server sends connected | Timer starts, waveform animates | P0 |
| Connected | Speak into mic | Audio echoes back (loopback) | P0 |
| Connected | Toggle mute | Audio stops sending, icon changes | P0 |
| Connected | Toggle speaker | Audio routes to loudspeaker | P1 |
| Connected | Press hang-up | Timer stops, transitions to ended | P0 |
| Ended | Shows duration | Correct MM:SS format | P0 |
| Ended | Press redial | Returns to calling with same number | P1 |
| Ended | Press back | Returns to dialer | P0 |
| Ended | Auto-return after 5s | Returns to dialer automatically | P1 |

---

## 10.3 Android Unit Tests

### CallTimer Test

```java
public class CallTimerTest {

    @Test
    public void testFormatTime() {
        assertEquals("00:00", CallTimer.formatTime(0));
        assertEquals("00:01", CallTimer.formatTime(1000));
        assertEquals("00:59", CallTimer.formatTime(59000));
        assertEquals("01:00", CallTimer.formatTime(60000));
        assertEquals("01:30", CallTimer.formatTime(90000));
        assertEquals("10:00", CallTimer.formatTime(600000));
        assertEquals("59:59", CallTimer.formatTime(3599000));
        assertEquals("01:00:00", CallTimer.formatTime(3600000));
    }

    @Test
    public void testFormatDuration() {
        assertEquals("0s", CallTimer.formatDuration(0));
        assertEquals("30s", CallTimer.formatDuration(30000));
        assertEquals("1m 30s", CallTimer.formatDuration(90000));
        assertEquals("5m 00s", CallTimer.formatDuration(300000));
        assertEquals("1h 00m 00s", CallTimer.formatDuration(3600000));
    }
}
```

### AgentPhonebook Test

```java
public class AgentPhonebookTest {

    @Test
    public void testLookupExisting() {
        assertEquals("Smile Dental Receptionist",
            AgentPhonebook.lookupAgent("101"));
        assertEquals("Billing Agent",
            AgentPhonebook.lookupAgent("102"));
    }

    @Test
    public void testLookupNonExistent() {
        assertNull(AgentPhonebook.lookupAgent("999"));
        assertNull(AgentPhonebook.lookupAgent(""));
    }

    @Test
    public void testHasNumber() {
        assertTrue(AgentPhonebook.hasNumber("101"));
        assertTrue(AgentPhonebook.hasNumber("102"));
        assertFalse(AgentPhonebook.hasNumber("999"));
    }
}
```

### TokenServerClient Test (Mock Mode)

```java
public class TokenServerClientTest {

    @Test
    public void testSimulateTokenSuccess() throws InterruptedException {
        final CountDownLatch latch = new CountDownLatch(1);
        final TokenServerClient.TokenResponse[] result = new TokenServerClient.TokenResponse[1];

        TokenServerClient client = new TokenServerClient("http://invalid-host:3000");
        client.simulateTokenRequest("101", new TokenServerClient.TokenCallback() {
            @Override
            public void onSuccess(TokenServerClient.TokenResponse response) {
                result[0] = response;
                latch.countDown();
            }

            @Override
            public void onError(String errorMessage) {
                fail("Should not error in mock mode");
            }
        });

        latch.await(3, TimeUnit.SECONDS);

        assertNotNull(result[0]);
        assertEquals("Smile Dental Receptionist", result[0].agentName);
        assertNotNull(result[0].token);
        assertNotNull(result[0].roomName);
    }

    @Test
    public void testSimulateTokenNotFound() throws InterruptedException {
        final CountDownLatch latch = new CountDownLatch(1);
        final String[] errorMessage = new String[1];

        TokenServerClient client = new TokenServerClient("http://invalid-host:3000");
        client.simulateTokenRequest("999", new TokenServerClient.TokenCallback() {
            @Override
            public void onSuccess(TokenServerClient.TokenResponse response) {
                fail("Should not succeed for invalid number");
            }

            @Override
            public void onError(String error) {
                errorMessage[0] = error;
                latch.countDown();
            }
        });

        latch.await(3, TimeUnit.SECONDS);
        assertNotNull(errorMessage[0]);
    }
}
```

---

## 10.4 Backend Tests

### Health Check Test

```bash
#!/bin/bash
# test-health.sh

echo "Testing health endpoint..."
RESPONSE=$(curl -s http://localhost:3000/health)
echo "Response: $RESPONSE"

# Check status field
STATUS=$(echo $RESPONSE | node -e "process.stdin.resume(); let d=''; process.stdin.on('data',c=>d+=c); process.stdin.on('end',()=>console.log(JSON.parse(d).status))")

if [ "$STATUS" = "ok" ]; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi
```

### Auth Endpoint Test

```bash
#!/bin/bash
# test-auth.sh

echo "Testing valid token..."
RESPONSE=$(curl -s -X POST http://localhost:3000/auth \
    -H "Content-Type: application/json" \
    -d '{"token":"user-token-123"}')
echo "Response: $RESPONSE"

echo "Testing invalid token..."
RESPONSE=$(curl -s -X POST http://localhost:3000/auth \
    -H "Content-Type: application/json" \
    -d '{"token":"invalid-token"}')
echo "Response: $RESPONSE"
```

### WebSocket Integration Test

```javascript
// test-websocket.js
const WebSocket = require('ws');

async function testCallFlow() {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket('ws://localhost:3000/call');
        const events = [];

        ws.on('open', () => {
            // Send dial message
            ws.send(JSON.stringify({
                action: 'dial',
                number: '101',
                token: 'user-token-123'
            }));
        });

        ws.on('message', (data) => {
            if (Buffer.isBuffer(data)) {
                // Audio data — ignore for this test
                return;
            }

            const msg = JSON.parse(data.toString());
            events.push(msg.event);

            if (msg.event === 'ringing') {
                console.log('✅ Received ringing');
            }

            if (msg.event === 'connected') {
                console.log('✅ Received connected');

                // Send test audio
                const testAudio = Buffer.alloc(640, 0);
                ws.send(testAudio, { binary: true });

                // Hang up
                ws.send(JSON.stringify({ action: 'hangup' }));
            }

            if (msg.event === 'ended') {
                console.log('✅ Received ended');
                ws.close();

                // Verify event sequence
                const expected = ['ringing', 'connected', 'ended'];
                const match = JSON.stringify(events) === JSON.stringify(expected);
                console.log(match ? '✅ Event sequence correct' : `❌ Expected ${expected}, got ${events}`);
                resolve(match);
            }
        });

        ws.on('error', (err) => {
            console.error('❌ WebSocket error:', err.message);
            reject(err);
        });

        // Timeout after 10 seconds
        setTimeout(() => {
            console.log('❌ Test timeout');
            ws.close();
            reject(new Error('Test timeout'));
        }, 10000);
    });
}

testCallFlow()
    .then(success => process.exit(success ? 0 : 1))
    .catch(err => { console.error(err); process.exit(1); });
```

---

## 10.5 End-to-End Test Procedure

### Manual E2E Test Script

```
┌─────────────────────────────────────────────────────────────┐
│              End-to-End Test Procedure                       │
│                                                             │
│  Prerequisites:                                             │
│  ☐ Backend running on ws://<IP>:3000                        │
│  ☐ Agent 101 running and registered                          │
│  ☐ Android app installed on device/emulator                  │
│  ☐ Device and backend on same network                       │
│                                                             │
│  Step 1: Basic Call                                         │
│  1. Open app → dialer screen appears                        │
│  2. Enter "101" on keypad                                   │
│  3. Press green call button                                 │
│  4. Verify: "Calling 101..." displayed                      │
│  5. Verify: Phone icon pulsing                              │
│  6. Verify: Transitions to "Ringing..."                     │
│  7. Verify: Agent name "Smile Dental Receptionist" shown    │
│  8. Verify: Transitions to "Connected"                      │
│  9. Verify: Timer starts at 00:00                           │
│  10. Verify: Waveform bars animating                        │
│                                                             │
│  Step 2: Audio Test                                         │
│  11. Speak into device microphone                           │
│  12. Verify: Audio echoes back through speaker (loopback)   │
│  13. Verify: Waveform responds to voice                     │
│  14. Verify: Audio is clear, no distortion                  │
│                                                             │
│  Step 3: Mute Test                                          │
│  15. Press mute button                                      │
│  16. Verify: Mute icon changes                              │
│  17. Speak into microphone                                  │
│  18. Verify: No audio echoes (muted)                        │
│  19. Press unmute button                                    │
│  20. Verify: Audio resumes                                  │
│                                                             │
│  Step 4: Hang-Up Test                                       │
│  21. Note timer value                                       │
│  22. Press red hang-up button                               │
│  23. Verify: Transitions to "Call Ended" screen             │
│  24. Verify: Duration matches noted timer value             │
│                                                             │
│  Step 5: Redial Test                                        │
│  25. Press "Call Again"                                     │
│  26. Verify: Returns to calling screen with "101"           │
│  27. Verify: Call reconnects                                │
│                                                             │
│  Step 6: Back to Dialer                                     │
│  28. Hang up again                                          │
│  29. Press "Back to Dialer"                                 │
│  30. Verify: Returns to dialer screen                       │
│  31. Verify: Number display is empty                        │
│                                                             │
│  Step 7: Invalid Number Test                                │
│  32. Enter "999" (non-existent)                             │
│  33. Press call                                             │
│  34. Verify: Error "Number not available" shown             │
│                                                             │
│  Step 8: Empty Number Test                                  │
│  35. Don't enter any number                                 │
│  36. Press call                                             │
│  37. Verify: Toast "Enter a number first"                   │
│                                                             │
│  Results: ___ / 37 tests passed                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 10.6 Performance Testing

### Latency Measurement

```
┌─────────────────────────────────────────────────────────────┐
│              Audio Latency Measurement                       │
│                                                             │
│  Method:                                                     │
│  1. Agent stub logs timestamp of first audio frame received │
│  2. Agent stub logs timestamp of first audio frame sent back│
│  3. Android app logs timestamp of audio playback start      │
│  4. Round-trip latency = T_playback - T_capture             │
│                                                             │
│  Target: < 200ms on local WiFi                              │
│  Acceptable: < 500ms                                        │
│  Unacceptable: > 1000ms                                     │
│                                                             │
│  Factors affecting latency:                                  │
│  • Network quality (WiFi signal strength)                   │
│  • Audio buffer size (640 bytes = 20ms)                     │
│  • WebSocket frame processing time                          │
│  • AudioRecord/AudioTrack buffer sizes                      │
│  • Device CPU performance                                   │
└─────────────────────────────────────────────────────────────┘
```

### Memory Testing

```bash
# Monitor Android app memory usage
adb shell dumpsys meminfo com.smiledental.dialer

# Monitor Node.js backend memory
# In server.js, add periodic logging:
setInterval(() => {
    const mem = process.memoryUsage();
    console.log(`Memory: RSS=${(mem.rss/1024/1024).toFixed(1)}MB`);
}, 60000);
```

### Stability Testing

| Test | Duration | Pass Criteria |
|------|----------|---------------|
| Continuous call | 30 minutes | No crashes, no memory growth |
| Rapid connect/disconnect | 100 cycles | No resource leaks |
| Network drop/reconnect | 10 cycles | Graceful recovery |
| Multiple simultaneous calls | 5 calls | All work independently |

---

*Next: [Page 11 — Error Handling](11-error-handling.md)*
