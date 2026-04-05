# Page 07 — Android API 16+ Compatibility

---

## 7.1 Overview

The app targets **Android 4.1 (Jelly Bean, API 16)** as the minimum SDK. This imposes significant constraints on available APIs, libraries, and language features. This page documents every compatibility consideration.

### Version Distribution (Historical Context)
| Android Version | API Level | Codename | Notes |
|----------------|-----------|----------|-------|
| 4.0 | 14-15 | Ice Cream Sandwich | Absolute minimum possible |
| **4.1-4.3** | **16-18** | **Jelly Bean** | **Our target minimum** |
| 4.4 | 19-20 | KitKat | Visualizer API available |
| 5.0 | 21-22 | Lollipop | ART runtime, Material Design |
| 6.0 | 23 | Marshmallow | Runtime permissions |

---

## 7.2 Java Language Level

### Constraint: Java 7 (source/target 1.7)

API 16 devices run Dalvik (not ART), which requires Java 7 bytecode.

| Feature | Available? | Notes |
|---------|-----------|-------|
| try-with-resources | ❌ | Java 7 feature, not supported on API < 19 |
| Diamond operator `<>` | ❌ | Java 7 feature, not supported on API < 19 |
| String in switch | ✅ | Java 7, works on API 16 |
| Multi-catch | ✅ | Java 7, works on API 16 |
| Lambda expressions | ❌ | Java 8, requires desugaring (not available) |
| Method references | ❌ | Java 8 |
| var keyword | ❌ | Java 10 |

### Implication for Code
```java
// ❌ DON'T — try-with-resources (Java 7+, API 19+)
try (FileInputStream fis = new FileInputStream(file)) {
    // ...
}

// ✅ DO — manual resource cleanup
FileInputStream fis = null;
try {
    fis = new FileInputStream(file);
    // ...
} finally {
    if (fis != null) {
        try { fis.close(); } catch (IOException e) { /* ignore */ }
    }
}

// ❌ DON'T — lambda (Java 8+)
button.setOnClickListener(v -> doSomething());

// ✅ DO — anonymous inner class
button.setOnClickListener(new View.OnClickListener() {
    @Override
    public void onClick(View v) {
        doSomething();
    }
});

// ❌ DON'T — diamond operator (Java 7+, API 19+)
Map<String, List<Integer>> map = new HashMap<>();

// ✅ DO — explicit type parameters
Map<String, List<Integer>> map = new HashMap<String, List<Integer>>();
```

---

## 7.3 Android SDK Compatibility

### Available APIs (API 16+)

| API | Available Since | Notes |
|-----|----------------|-------|
| `AudioRecord` | API 1 | ✅ Fully available |
| `AudioTrack` | API 1 | ✅ Fully available |
| `ToneGenerator` | API 1 | ✅ Fully available |
| `AudioManager` | API 1 | ✅ Fully available |
| `Visualizer` | API 19 | ⚠️ API 19+ only — need fallback |
| `MediaPlayer` | API 1 | ✅ Fully available |
| `Handler` | API 1 | ✅ Fully available |
| `AsyncTask` | API 3 | ✅ Available but deprecated — use Thread |
| `Thread` | API 1 | ✅ Preferred over AsyncTask |
| `JSONObject` | API 1 | ✅ Fully available |
| `OkHttp` | — | ✅ 3.12.x supports API 16 |

### APIs NOT Available on API 16

| API | Minimum API | Workaround |
|-----|------------|------------|
| `Visualizer` | 19 | Simulated waveform with random animation |
| `AudioRecord.Builder` | 23 | Use constructor with parameters directly |
| `AudioTrack.Builder` | 23 | Use constructor with parameters directly |
| `MediaRecorder.AudioSource.VOICE_COMMUNICATION` | 11 | Use `MIC` instead |
| `Context.checkSelfPermission` | 23 | Use `ContextCompat.checkSelfPermission` |
| `Notification.Builder` (full) | 16 | Use `NotificationCompat` from support library |
| `VectorDrawable` | 21 | Use PNG drawables or `appcompat` backport |

---

## 7.4 Permission Model

### API 16-22: Install-Time Permissions

On Android 4.1-5.1, **all permissions are granted at install time**. No runtime request needed.

```xml
<!-- These are granted when user installs the app -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.RECORD_AUDIO" />
<uses-permission android:name="android.permission.MODIFY_AUDIO_SETTINGS" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

### API 23+: Runtime Permissions

On Android 6.0+, `RECORD_AUDIO` is a **dangerous permission** requiring runtime request.

```java
// Check permission
boolean hasPermission = ContextCompat.checkSelfPermission(this,
    Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED;

if (!hasPermission) {
    ActivityCompat.requestPermissions(this,
        new String[]{Manifest.permission.RECORD_AUDIO},
        REQUEST_CODE_MIC);
}
```

### Compatibility Pattern
```java
private boolean checkAndRequestMicPermission() {
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        // Runtime permission check (API 23+)
        if (ContextCompat.checkSelfPermission(this,
                Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(this,
                new String[]{Manifest.permission.RECORD_AUDIO},
                REQUEST_CODE_MIC);
            return false;  // Permission not yet granted
        }
    }
    // API 16-22: permission granted at install time
    return true;
}
```

---

## 7.5 OkHttp Compatibility

### Version Selection

| OkHttp Version | Min API | Notes |
|---------------|---------|-------|
| 3.12.x | API 16+ (Java 7+) | ✅ **Our choice** — last API 16 compatible |
| 4.x | API 21+ (Java 8+) | ❌ Requires Kotlin stdlib, API 21+ |
| 2.x | API 9+ | ❌ Too old, no WebSocket support |

### Dependency
```groovy
dependencies {
    implementation 'com.squareup.okhttp3:okhttp:3.12.13'
}
```

### WebSocket Support
OkHttp 3.12.x includes WebSocket support via `WebSocket` and `WebSocketListener` classes.

```java
OkHttpClient client = new OkHttpClient.Builder()
    .readTimeout(0, TimeUnit.MILLISECONDS)  // No timeout
    .pingInterval(30, TimeUnit.SECONDS)     // Keep-alive pings
    .build();

Request request = new Request.Builder()
    .url("ws://192.168.1.100:3000/call")
    .build();

WebSocket ws = client.newWebSocket(request, new WebSocketListener() {
    @Override
    public void onOpen(WebSocket webSocket, Response response) {
        // Connection established
    }

    @Override
    public void onMessage(WebSocket webSocket, String text) {
        // Text message (JSON)
    }

    @Override
    public void onMessage(WebSocket webSocket, ByteString bytes) {
        // Binary message (PCM audio)
    }

    @Override
    public void onFailure(WebSocket webSocket, Throwable t, Response response) {
        // Connection failed
    }

    @Override
    public void onClosed(WebSocket webSocket, int code, String reason) {
        // Connection closed
    }
});
```

---

## 7.6 UI Compatibility

### Layout Considerations

| Feature | API 16 Support | Notes |
|---------|---------------|-------|
| `GridLayout` | API 14+ (via v7 support) | Use `androidx.gridlayout:gridlayout:1.0.0` |
| `CardView` | API 14+ (via v7 support) | Use `androidx.cardview:cardview:1.0.0` |
| `VectorDrawable` | API 21+ native | Use PNG or `appcompat` backport |
| `ConstraintLayout` | API 9+ (via support lib) | Use `androidx.constraintlayout:constraintlayout:1.1.3` |
| `ObjectAnimator` | API 16+ | ✅ Available — use for pulse animation |
| `ViewPropertyAnimator` | API 16+ | ✅ Available |

### Density Independence

All dimensions should use `dp` (density-independent pixels) and `sp` (scale-independent pixels for text):

```xml
<!-- ✅ Good -->
android:layout_width="72dp"
android:textSize="28sp"

<!-- ❌ Bad — will look different on different screens -->
android:layout_width="72px"
android:textSize="28px"
```

### Screen Size Support

| Screen Size | Resolution | Density | Notes |
|-------------|-----------|---------|-------|
| Small | 240×320 | ldpi | Rare on API 16+ |
| Normal | 320×480 | mdpi | Minimum target |
| Large | 480×800 | hdpi | Common |
| xLarge | 720×1280 | xhdpi | Modern phones |

The dialer UI uses `GridLayout` with proportional sizing to adapt to all screen sizes.

---

## 7.7 Threading Considerations

### No AsyncTask

`AsyncTask` is deprecated and has known issues on older Android versions (thread pool changes, memory leaks). Use plain `Thread` instead:

```java
// ✅ Preferred pattern
private volatile boolean running = false;
private Thread workerThread;

void startWork() {
    running = true;
    workerThread = new Thread(new Runnable() {
        @Override
        public void run() {
            while (running) {
                // Do work
            }
        }
    }, "Worker-Thread");
    workerThread.setPriority(Thread.MAX_PRIORITY);
    workerThread.start();
}

void stopWork() {
    running = false;
    if (workerThread != null) {
        workerThread.interrupt();
        try {
            workerThread.join(1000);
        } catch (InterruptedException e) {
            workerThread.interrupt();
        }
    }
}
```

### UI Thread Updates

All View modifications must happen on the main (UI) thread:

```java
// From background thread → update UI
runOnUiThread(new Runnable() {
    @Override
    public void run() {
        timerText.setText(formattedTime);
    }
});

// Or use View.post()
waveformView.post(new Runnable() {
    @Override
    public void run() {
        waveformView.invalidate();
    }
});
```

---

## 7.8 Memory Constraints

### Heap Size Limits

| Device Class | Heap Limit | Notes |
|-------------|-----------|-------|
| Low-end (API 16) | 16-32 MB | Very constrained |
| Mid-range | 64-128 MB | Reasonable |
| High-end | 256-512 MB | Plenty |

### Memory Best Practices
- Avoid large object allocations in audio thread (reuse buffers)
- Release AudioRecord/AudioTrack promptly on hangup
- Don't hold references to destroyed activities
- Use `WeakReference` for callbacks if needed

```java
// ✅ Reuse audio buffer
private byte[] audioBuffer = new byte[640];

void readAudio() {
    // Reuse same buffer each iteration
    int bytesRead = audioRecord.read(audioBuffer, 0, audioBuffer.length);
}

// ❌ Don't allocate new buffer each iteration
void readAudio() {
    byte[] buffer = new byte[640];  // GC pressure!
    audioRecord.read(buffer, 0, buffer.length);
}
```

---

*Next: [Page 08 — Security Design](08-security-design.md)*
