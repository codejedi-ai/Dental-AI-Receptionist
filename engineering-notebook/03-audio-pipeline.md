# Page 03 — Audio Pipeline

---

## 3.1 Overview

The audio pipeline is the core of the system — it captures microphone input, streams it to the backend via WebSocket, receives the agent's audio response, and plays it back through the speaker — all in real-time.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Android Device                               │
│                                                                     │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐          │
│  │Microphone│───►│  AudioRecord  │───►│  WebSocketManager│          │
│  │  (HW)    │    │  (PCM 16bit)  │    │  (binary frames) │          │
│  └──────────┘    └──────────────┘    └────────┬─────────┘          │
│                                                │                    │
│                                                │ WebSocket          │
│                                                │ ws://backend:3000  │
│                                                │                    │
│  ┌──────────┐    ┌──────────────┐    ┌────────┴─────────┐          │
│  │ Speaker  │◄───│  AudioTrack   │◄───│  WebSocketManager│          │
│  │  (HW)    │    │  (PCM 16bit)  │    │  (binary frames) │          │
│  └──────────┘    └──────────────┘    └──────────────────┘          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3.2 Audio Format Specification

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Encoding | PCM 16-bit | Standard telephony quality, no codec overhead |
| Channels | Mono | Single microphone, single speaker |
| Sample Rate | 16,000 Hz | Good voice quality, low bandwidth |
| Bit Rate | 256 kbps (16000 × 16) | ~32 KB/s per direction |
| Frame Size | 640 bytes (320 samples) | 20ms chunks, low latency |
| Buffer Size | `getMinBufferSize()` × 2 | Prevents underrun |

### Bandwidth Calculation
```
16,000 samples/sec × 2 bytes/sample = 32,000 bytes/sec = 256 kbps
Per 20ms frame: 640 bytes
WebSocket overhead: ~6 bytes per frame
Total: ~32 KB/s × 2 (full-duplex) = 64 KB/s
```

This is well within local WiFi capacity.

---

## 3.3 AudioRecord Setup

### Initialization
```java
int sampleRate = 16000;
int channelConfig = AudioFormat.CHANNEL_IN_MONO;
int audioFormat = AudioFormat.ENCODING_PCM_16BIT;

int bufferSize = AudioRecord.getMinBufferSize(sampleRate, channelConfig, audioFormat);
bufferSize = bufferSize * 2;  // Double for safety margin

AudioRecord audioRecord = new AudioRecord(
    MediaRecorder.AudioSource.MIC,
    sampleRate,
    channelConfig,
    audioFormat,
    bufferSize
);

if (audioRecord.getState() != AudioRecord.STATE_INITIALIZED) {
    // Handle initialization failure
    throw new RuntimeException("AudioRecord failed to initialize");
}
```

### Recording Thread
```java
private volatile boolean isRecording = false;
private Thread recordingThread;

void startRecording() {
    isRecording = true;
    audioRecord.startRecording();

    recordingThread = new Thread(new Runnable() {
        @Override
        public void run() {
            byte[] buffer = new byte[640];  // 20ms at 16kHz, 16-bit mono

            while (isRecording) {
                int bytesRead = audioRecord.read(buffer, 0, buffer.length);

                if (bytesRead > 0 && !isMuted) {
                    // Send PCM data via WebSocket
                    webSocketManager.sendAudio(buffer, bytesRead);

                    // Update waveform visualizer
                    float amplitude = calculateAmplitude(buffer, bytesRead);
                    waveformView.updateAmplitude(amplitude);
                }
            }
        }
    }, "AudioRecord-Thread");
    recordingThread.setPriority(Thread.MAX_PRIORITY);
    recordingThread.start();
}

void stopRecording() {
    isRecording = false;
    if (recordingThread != null) {
        try {
            recordingThread.join(1000);  // Wait up to 1 second
        } catch (InterruptedException e) {
            recordingThread.interrupt();
        }
    }
    audioRecord.stop();
    audioRecord.release();
}
```

---

## 3.4 AudioTrack Setup

### Initialization
```java
int sampleRate = 16000;
int channelConfig = AudioFormat.CHANNEL_OUT_MONO;
int audioFormat = AudioFormat.ENCODING_PCM_16BIT;

int bufferSize = AudioTrack.getMinBufferSize(sampleRate, channelConfig, audioFormat);
bufferSize = bufferSize * 2;

AudioTrack audioTrack = new AudioTrack(
    AudioManager.STREAM_VOICE_CALL,
    sampleRate,
    channelConfig,
    audioFormat,
    bufferSize,
    AudioTrack.MODE_STREAM
);

if (audioTrack.getState() != AudioTrack.STATE_INITIALIZED) {
    throw new RuntimeException("AudioTrack failed to initialize");
}
```

### Playback
```java
void startPlayback() {
    audioTrack.play();
}

// Called from WebSocketManager when audio data arrives
void playAudioData(byte[] data, int length) {
    if (audioTrack != null && audioTrack.getPlayState() == AudioTrack.PLAYSTATE_PLAYING) {
        audioTrack.write(data, 0, length);
    }
}

void stopPlayback() {
    if (audioTrack != null) {
        audioTrack.stop();
        audioTrack.flush();
        audioTrack.release();
        audioTrack = null;
    }
}
```

---

## 3.5 AudioEngine Wrapper Class

Combines AudioRecord and AudioTrack into a single manageable interface:

```java
public class AudioEngine {

    private AudioRecord audioRecord;
    private AudioTrack audioTrack;
    private WebSocketManager wsManager;
    private WaveformView waveformView;

    private volatile boolean isRecording = false;
    private volatile boolean isMuted = false;
    private Thread recordingThread;

    private int sampleRate = 16000;
    private int bufferSize;

    public AudioEngine(Context context, WebSocketManager wsManager, WaveformView waveformView) {
        this.wsManager = wsManager;
        this.waveformView = waveformView;
        initAudioRecord();
        initAudioTrack();
    }

    private void initAudioRecord() {
        int minBuffer = AudioRecord.getMinBufferSize(
            sampleRate,
            AudioFormat.CHANNEL_IN_MONO,
            AudioFormat.ENCODING_PCM_16BIT
        );
        bufferSize = minBuffer * 2;

        audioRecord = new AudioRecord(
            MediaRecorder.AudioSource.MIC,
            sampleRate,
            AudioFormat.CHANNEL_IN_MONO,
            AudioFormat.ENCODING_PCM_16BIT,
            bufferSize
        );
    }

    private void initAudioTrack() {
        int minBuffer = AudioTrack.getMinBufferSize(
            sampleRate,
            AudioFormat.CHANNEL_OUT_MONO,
            AudioFormat.ENCODING_PCM_16BIT
        );

        audioTrack = new AudioTrack(
            AudioManager.STREAM_VOICE_CALL,
            sampleRate,
            AudioFormat.CHANNEL_OUT_MONO,
            AudioFormat.ENCODING_PCM_16BIT,
            minBuffer * 2,
            AudioTrack.MODE_STREAM
        );
    }

    public void start() {
        // Start recording and playback
        isRecording = true;
        audioTrack.play();
        audioRecord.startRecording();

        recordingThread = new Thread(new Runnable() {
            @Override
            public void run() {
                byte[] buffer = new byte[640];

                while (isRecording) {
                    int bytesRead = audioRecord.read(buffer, 0, buffer.length);

                    if (bytesRead > 0 && !isMuted) {
                        wsManager.sendAudio(buffer, bytesRead);
                    }

                    // Update waveform
                    float amp = calculateAmplitude(buffer, bytesRead);
                    waveformView.postUpdateAmplitude(amp);
                }
            }
        }, "AudioEngine-Thread");
        recordingThread.setPriority(Thread.MAX_PRIORITY);
        recordingThread.start();
    }

    public void stop() {
        isRecording = false;

        if (recordingThread != null) {
            try {
                recordingThread.join(1000);
            } catch (InterruptedException e) {
                recordingThread.interrupt();
            }
        }

        if (audioRecord != null) {
            audioRecord.stop();
            audioRecord.release();
        }

        if (audioTrack != null) {
            audioTrack.stop();
            audioTrack.flush();
            audioTrack.release();
        }
    }

    public void setMuted(boolean muted) {
        this.isMuted = muted;
    }

    public boolean isMuted() {
        return isMuted;
    }

    private float calculateAmplitude(byte[] buffer, int length) {
        float sum = 0;
        for (int i = 0; i < length; i += 2) {
            if (i + 1 < length) {
                short sample = (short) ((buffer[i + 1] << 8) | (buffer[i] & 0xFF));
                sum += Math.abs(sample);
            }
        }
        int sampleCount = length / 2;
        return sampleCount > 0 ? (sum / sampleCount) / 32768f : 0;
    }
}
```

---

## 3.6 Mute and Speaker Controls

### Mute (Microphone)
```java
// Software mute — stop sending audio to WebSocket
public void setMuted(boolean muted) {
    this.isMuted = muted;
    // Note: AudioRecord continues running, we just don't send the data
}
```

### Speaker (AudioManager)
```java
AudioManager audioManager = (AudioManager) context.getSystemService(Context.AUDIO_SERVICE);

// Enable speakerphone (loudspeaker)
public void setSpeakerphoneOn(boolean on) {
    audioManager.setSpeakerphoneOn(on);
    audioManager.setMode(on ? AudioManager.MODE_IN_COMMUNICATION : AudioManager.MODE_NORMAL);
}
```

---

## 3.7 Threading Model

```
┌─────────────────────────────────────────────────────────────┐
│                      Main Thread (UI)                       │
│  • Activity lifecycle                                       │
│  • Button clicks                                            │
│  • View updates (timer, waveform)                           │
│  • State transitions                                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  AudioRecord Thread                         │
│  • Continuous PCM capture                                   │
│  • Buffer read → WebSocket send                             │
│  • Amplitude calculation → waveform update (post to UI)     │
│  Priority: MAX_PRIORITY                                     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│               WebSocket Thread (OkHttp internal)            │
│  • JSON message receive → state transition (post to UI)     │
│  • Binary audio frame receive → AudioTrack write            │
│  • Connection state management                              │
└─────────────────────────────────────────────────────────────┘
```

### Thread Safety Notes
- `isRecording` and `isMuted` are `volatile` for cross-thread visibility
- Waveform updates use `View.post()` to cross from audio thread to UI thread
- State transitions use `Activity.runOnUiThread()` from WebSocket callback
- No shared mutable state between threads except atomic/volatile flags

---

## 3.8 API 16-18 Compatibility

The `Visualizer` class (for real audio amplitude capture) was added in API 19. For API 16-18:

```java
if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
    // Use real Visualizer
    visualizer = new Visualizer(audioSessionId);
    visualizer.setDataCaptureListener(new Visualizer.OnDataCaptureListener() {
        @Override
        public void onWaveFormDataCapture(Visualizer v, byte[] waveform, int samplingRate) {
            // Convert waveform to amplitude values
            float[] amplitudes = convertToAmplitudes(waveform);
            waveformView.updateAmplitudes(amplitudes);
        }
    }, Visualizer.getMaxCaptureRate(), true, false);
    visualizer.setEnabled(true);
} else {
    // Simulated waveform — random amplitude animation
    waveformView.setSimulatedMode(true);
}
```

---

*Next: [Page 04 — WebSocket Protocol](04-websocket-protocol.md)*
