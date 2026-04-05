# Page 02 — Android App Design

---

## 2.1 Project Structure

```
android-app/
├── app/
│   ├── build.gradle
│   ├── src/main/
│   │   ├── AndroidManifest.xml
│   │   ├── java/com/smiledental/dialer/
│   │   │   ├── DialerActivity.java       # Keypad screen (launcher)
│   │   │   ├── CallActivity.java          # Call screen (all states)
│   │   │   ├── CallState.java             # Enum: DIALING, RINGING, CONNECTED, ENDED
│   │   │   ├── WebSocketManager.java      # OkHttp WS + audio send/receive
│   │   │   ├── AudioEngine.java           # AudioRecord + AudioTrack wrapper
│   │   │   └── WaveformView.java          # Custom View: audio amplitude bars
│   │   ├── res/
│   │   │   ├── layout/
│   │   │   │   ├── activity_dialer.xml    # Dialer screen layout
│   │   │   │   └── activity_call.xml      # Call screen layout (all states)
│   │   │   ├── drawable/
│   │   │   │   ├── btn_call.xml           # Green circle selector
│   │   │   │   ├── btn_hangup.xml         # Red circle selector
│   │   │   │   ├── btn_key.xml            # Keypad button selector
│   │   │   │   └── bg_gradient.xml        # Background gradient
│   │   │   ├── anim/
│   │   │   │   └── pulse.xml              # Phone icon pulse animation
│   │   │   ├── values/
│   │   │   │   ├── config.xml             # Backend URL, auth token
│   │   │   │   ├── colors.xml             # Color definitions
│   │   │   │   ├── strings.xml            # String resources
│   │   │   │   └── styles.xml             # Theme definitions
│   │   │   └── raw/
│   │   │       └── ring.ogg               # Ringtone (generated tone)
│   └── proguard-rules.pro
├── build.gradle (project-level)
├── settings.gradle
└── gradle.properties
```

---

## 2.2 Screen 1: DialerActivity

### Layout Design (`activity_dialer.xml`)

```
┌─────────────────────────────────────┐
│         Smile Dental 🦷             │  ← TextView, centered, 18sp
│                                     │
│  ┌───────────────────────────────┐  │
│  │                               │  │  ← Number display field
│  │        1 0 1                  │  │    Monospace, 36sp, right-aligned
│  │                               │  │    Dark background, light text
│  └───────────────────────────────┘  │
│                                     │
│  ┌───────┐ ┌───────┐ ┌───────┐     │
│  │   1   │ │   2   │ │   3   │     │  ← Keypad row 1
│  └───────┘ └───────┘ └───────┘     │    72dp × 72dp circular buttons
│  ┌───────┐ ┌───────┐ ┌───────┐     │    Dark gray background, white text
│  │   4   │ │   5   │ │   6   │     │    DTMF tone on press
│  └───────┘ └───────┘ └───────┘     │
│  ┌───────┐ ┌───────┐ ┌───────┐     │
│  │   7   │ │   8   │ │   9   │     │  ← Keypad row 3
│  └───────┘ └───────┘ └───────┘     │
│  ┌───────┐ ┌───────┐ ┌───────┐     │
│  │   *   │ │   0   │ │   #   │     │  ← Keypad row 4
│  └───────┘ └───────┘ └───────┘     │    Long-press 0 for "+"
│                                     │
│  ┌─────────────────────────────┐   │
│  │         ⌫        📞         │   │  ← Backspace + Call button
│  │      (gray)    (green)      │   │    56dp circular buttons
│  └─────────────────────────────┘   │
│                                     │
└─────────────────────────────────────┘
```

### Color Scheme
| Element | Color | Hex |
|---------|-------|-----|
| Background | Dark charcoal | `#1a1a2e` |
| Display field | Darker charcoal | `#16213e` |
| Display text | White | `#ffffff` |
| Keypad buttons | Medium gray | `#2c2c3e` |
| Keypad text | White | `#ffffff` |
| Keypad pressed | Lighter gray | `#3a3a4e` |
| Call button | Green | `#4CAF50` |
| Call button pressed | Dark green | `#388E3C` |
| Backspace button | Gray | `#607D8B` |
| Clinic name | Light blue | `#64B5F6` |

### Behavior
| Action | Result |
|--------|--------|
| Key press | Append digit to display, play DTMF tone via `ToneGenerator` |
| Backspace press | Remove last digit from display |
| Backspace long-press | Clear entire display |
| 0 long-press | Insert "+" character |
| Call button press | Validate number not empty → start CallActivity with number extra |
| Empty number + call | Show toast "Enter a number first" |

### DTMF Tone Generation
```java
// Using Android's built-in ToneGenerator
ToneGenerator toneGen = new ToneGenerator(
    AudioManager.STREAM_DTMF,
    ToneGenerator.TONE_DTMF_100MS
);

// Map each key to its DTMF tone
void playDTMF(char digit) {
    switch(digit) {
        case '0': toneGen.startTone(ToneGenerator.TONE_DTMF_0); break;
        case '1': toneGen.startTone(ToneGenerator.TONE_DTMF_1); break;
        // ... etc for all digits
        case '*': toneGen.startTone(ToneGenerator.TONE_DTMF_S); break;
        case '#': toneGen.startTone(ToneGenerator.TONE_DTMF_P); break;
    }
}
```

---

## 2.3 Screen 2: CallActivity

### Single Activity, Multiple States

Rather than separate activities for each call state, a **single CallActivity** manages all states via a state machine. This avoids activity lifecycle complexity during state transitions.

### Layout Design (`activity_call.xml`)

The layout contains all state views stacked on top of each other, with visibility toggled based on current state:

```
┌─────────────────────────────────────┐
│                                     │
│  ┌─────────────────────────────┐   │
│  │     DIALING STATE VIEW      │   │  ← Visible during DIALING
│  │                             │   │
│  │        📞 (pulsing)         │   │  ← Phone icon with ObjectAnimator
│  │                             │   │
│  │     Calling 101...          │   │  ← "Calling <number>..."
│  │                             │   │
│  │     [ Cancel ]              │   │  ← Red cancel button
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │     RINGING STATE VIEW      │   │  ← Visible during RINGING
│  │                             │   │
│  │        📞 (pulsing)         │   │  ← Phone icon with ObjectAnimator
│  │                             │   │
│  │     Ringing...              │   │  ← Animated dots "Ringing."
│  │     Smile Dental Recept.    │   │  ← Agent name (if available)
│  │                             │   │
│  │     [ Cancel ]              │   │  ← Red cancel button
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │    CONNECTED STATE VIEW     │   │  ← Visible during CONNECTED
│  │                             │   │
│  │     Smile Dental Recept.    │   │  ← Agent name
│  │     02:34                   │   │  ← Call timer (MM:SS, monospace)
│  │                             │   │
│  │  ┌─┬ ┬ ┬─┬ ┬ ┬─┬ ┬ ┬─┐    │   │  ← WaveformView (5-7 bars)
│  │  │ │││ │││ │││ │││ │││    │   │
│  │  └─┴─┴─┴─┴─┴─┴─┴─┴─┘    │   │
│  │                             │   │
│  │  [🔇]          [📢]        │   │  ← Mute + Speaker toggles
│  │                             │   │
│  │        [ 📞 Hang Up ]       │   │  ← Large red hang-up button
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │      ENDED STATE VIEW       │   │  ← Visible during ENDED
│  │                             │   │
│  │       Call Ended            │   │
│  │       Duration: 2m 34s      │   │
│  │                             │   │
│  │    [ Call Again ]           │   │  ← Green button → redial
│  │    [ Back to Dialer ]       │   │  ← Gray button → dialer
│  └─────────────────────────────┘   │
│                                     │
└─────────────────────────────────────┘
```

### State Machine

```
                    ┌─────────┐
                    │ DIALING │
                    └────┬────┘
                         │
              {ringing}  │
              from server│
                         │
                    ┌────┴────┐
                    │ RINGING │
                    └────┬────┘
                         │
              {connected}│
              from server│
                         │
                   ┌─────┴──────┐
                   │ CONNECTED  │
                   └─────┬──────┘
                         │
        {ended} from server OR hangup pressed
                         │
                    ┌────┴────┐
                    │  ENDED  │
                    └─────────┘
                         │
              Auto-return after 5s
              OR user action
                         │
                    ┌────┴────┐
                    │ DIALER  │
                    └─────────┘
```

### State Transitions Table

| From State | Trigger | To State | Actions |
|-----------|---------|----------|---------|
| DIALING | Server sends `ringing` | RINGING | Show agent name, start ring sound |
| DIALING | Server sends `busy` | ENDED | Show "Number not available" |
| DIALING | User presses Cancel | DIALER | Close WS, finish activity |
| DIALING | Timeout (30s) | ENDED | Show "Connection timeout" |
| RINGING | Server sends `connected` | CONNECTED | Stop ring sound, start timer, start audio |
| RINGING | Server sends `busy` | ENDED | Show "Agent unavailable" |
| RINGING | User presses Cancel | DIALER | Send hangup, close WS, finish |
| RINGING | Timeout (60s) | ENDED | Show "No answer" |
| CONNECTED | Server sends `ended` | ENDED | Stop audio, stop timer |
| CONNECTED | User presses Hang Up | ENDED | Send hangup, stop audio, stop timer |
| ENDED | Auto-timer (5s) | DIALER | Finish activity |
| ENDED | User presses Call Again | DIALING | Restart with same number |
| ENDED | User presses Back | DIALER | Finish activity |

---

## 2.4 Custom Views

### WaveformView

A custom `View` that draws 5-7 vertical bars representing audio amplitude.

**API 19+ (KitKat):** Uses `android.media.audiofx.Visualizer` to capture actual audio output amplitude.

**API 16-18 (Jelly Bean):** Simulated bars with random amplitude animation (no Visualizer API available).

```java
public class WaveformView extends View {
    private Paint barPaint;
    private float[] barHeights;  // Normalized 0.0 - 1.0
    private int barCount = 7;
    private boolean useSimulated = true;  // true for API < 19

    @Override
    protected void onDraw(Canvas canvas) {
        float barWidth = getWidth() / barCount;
        float maxHeight = getHeight() * 0.8f;

        for (int i = 0; i < barCount; i++) {
            float barHeight = barHeights[i] * maxHeight;
            float left = i * barWidth + 2;
            float right = (i + 1) * barWidth - 2;
            float top = (getHeight() - barHeight) / 2;
            float bottom = top + barHeight;

            canvas.drawRect(left, top, right, bottom, barPaint);
        }
    }

    // Called from AudioEngine when new audio data arrives
    public void updateAmplitudes(float[] values) {
        barHeights = values;
        invalidate();  // Triggers onDraw
    }
}
```

---

## 2.5 Drawable Resources

### btn_call.xml (Green Circle)
```xml
<?xml version="1.0" encoding="utf-8"?>
<selector xmlns:android="http://schemas.android.com/apk/res/android">
    <item android:state_pressed="true">
        <shape android:shape="oval">
            <solid android:color="#388E3C" />
        </shape>
    </item>
    <item>
        <shape android:shape="oval">
            <solid android:color="#4CAF50" />
        </shape>
    </item>
</selector>
```

### btn_hangup.xml (Red Circle)
```xml
<?xml version="1.0" encoding="utf-8"?>
<selector xmlns:android="http://schemas.android.com/apk/res/android">
    <item android:state_pressed="true">
        <shape android:shape="oval">
            <solid android:color="#D32F2F" />
        </shape>
    </item>
    <item>
        <shape android:shape="oval">
            <solid android:color="#F44336" />
        </shape>
    </item>
</selector>
```

### bg_gradient.xml (Background)
```xml
<?xml version="1.0" encoding="utf-8"?>
<shape xmlns:android="http://schemas.android.com/apk/res/android">
    <gradient
        android:angle="270"
        android:startColor="#1a1a2e"
        android:endColor="#0f0f1a"
        android:type="linear" />
</shape>
```

---

## 2.6 Animation Resources

### pulse.xml (Phone Icon Animation)
```xml
<?xml version="1.0" encoding="utf-8"?>
<set xmlns:android="http://schemas.android.com/apk/res/android"
    android:interpolator="@android:anim/accelerate_decelerate_interpolator">
    <scale
        android:duration="1000"
        android:fromXScale="1.0"
        android:fromYScale="1.0"
        android:toXScale="1.2"
        android:toYScale="1.2"
        android:pivotX="50%"
        android:pivotY="50%"
        android:repeatCount="infinite"
        android:repeatMode="reverse" />
    <alpha
        android:duration="1000"
        android:fromAlpha="1.0"
        android:toAlpha="0.5"
        android:repeatCount="infinite"
        android:repeatMode="reverse" />
</set>
```

---

## 2.7 Configuration Resources

### config.xml
```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="backend_ws_url">ws://192.168.1.100:3000/call</string>
    <string name="auth_token">user-token-123</string>
    <integer name="call_timeout_dialing_ms">30000</integer>
    <integer name="call_timeout_ringing_ms">60000</integer>
    <integer name="call_ended_auto_return_ms">5000</integer>
</resources>
```

---

*Next: [Page 03 — Audio Pipeline](03-audio-pipeline.md)*
