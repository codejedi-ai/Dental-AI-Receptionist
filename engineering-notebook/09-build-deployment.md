# Page 09 — Build & Deployment

---

## 9.1 Overview

This page covers building the Android APK, setting up the backend server, and deploying the complete system.

---

## 9.2 Android App Build

### Prerequisites

| Requirement | Version | Notes |
|------------|---------|-------|
| Android Studio | 3.6.3+ | Or compatible IDE |
| JDK | 8-11 | JDK 21 may require AGP 8.7+ |
| Android SDK | API 34 | Build tools 34.0.0 |
| Gradle | 8.5+ | Via Gradle wrapper |
| Android Gradle Plugin | 8.2+ | In project build.gradle |

### Project-Level build.gradle

```groovy
// Top-level build file
buildscript {
    repositories {
        google()
        mavenCentral()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:8.2.0'
    }
}

// Note: repositories managed in settings.gradle for AGP 8+
```

### settings.gradle

```groovy
pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}
rootProject.name = 'SmileDental-Dialer'
include ':app'
```

### App-level build.gradle

```groovy
apply plugin: 'com.android.application'

android {
    namespace 'com.smiledental.dialer'
    compileSdkVersion 34

    defaultConfig {
        applicationId "com.smiledental.dialer"
        minSdkVersion 16
        targetSdkVersion 34
        versionCode 1
        versionName "1.0"

        testInstrumentationRunner "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        release {
            minifyEnabled false
            proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'),
                          'proguard-rules.pro'
        }
    }

    compileOptions {
        sourceCompatibility JavaVersion.VERSION_1_7
        targetCompatibility JavaVersion.VERSION_1_7
    }
}

dependencies {
    implementation 'com.squareup.okhttp3:okhttp:3.12.13'
    implementation 'androidx.appcompat:appcompat:1.3.1'
    implementation 'androidx.constraintlayout:constraintlayout:1.1.3'

    testImplementation 'junit:junit:4.13.2'
    androidTestImplementation 'androidx.test.ext:junit:1.1.5'
    androidTestImplementation 'androidx.test.espresso:espresso-core:3.5.1'
}
```

### gradle.properties

```properties
org.gradle.jvmargs=-Xmx2048m -Dfile.encoding=UTF-8
android.useAndroidX=true
android.enableJetifier=true
android.nonTransitiveRClass=true
```

### Build Commands

```bash
# Navigate to project root
cd android-app/

# Clean build
./gradlew clean

# Build debug APK
./gradlew assembleDebug

# Build release APK (requires signing)
./gradlew assembleRelease

# Run tests
./gradlew test

# Install on connected device
./gradlew installDebug
```

### APK Output Locations

| Build Type | Output Path |
|-----------|-------------|
| Debug | `app/build/outputs/apk/debug/app-debug.apk` |
| Release | `app/build/outputs/apk/release/app-release.apk` |

### APK Signing (Release)

```groovy
// In app/build.gradle
android {
    signingConfigs {
        release {
            storeFile file("../keystore/smiledental-release.jks")
            storePassword System.getenv("KEYSTORE_PASSWORD")
            keyAlias "smiledental"
            keyPassword System.getenv("KEY_PASSWORD")
        }
    }

    buildTypes {
        release {
            signingConfig signingConfigs.release
            minifyEnabled true
            proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'),
                          'proguard-rules.pro'
        }
    }
}
```

Generate keystore:
```bash
keytool -genkey -v \
    -keystore keystore/smiledental-release.jks \
    -keyalg RSA \
    -keysize 2048 \
    -validity 10000 \
    -alias smiledental
```

---

## 9.3 Backend Server Setup

### Prerequisites

| Requirement | Version |
|------------|---------|
| Node.js | 14.0+ |
| npm | 6.0+ |

### package.json

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

### .env

```env
PORT=3000
```

### phonebook.json

```json
{
    "101": "Smile Dental Receptionist",
    "102": "Billing Agent",
    "103": "Appointment Reminder Agent",
    "104": "Patient Support Agent",
    "105": "Emergency Line"
}
```

### Start Backend

```bash
cd backend/

# Install dependencies (first time only)
npm install

# Start server
npm start

# Expected output:
# 🦷 Smile Dental Backend Server
#    Listening on ws://0.0.0.0:3000
#    Phonebook: 5 agents
#    Valid tokens: 2
```

### Verify Backend

```bash
# Health check
curl http://localhost:3000/health

# Expected response:
# {"status":"ok","agents":[],"activeCalls":0}
```

---

## 9.4 Agent Stub Setup

### package.json

```json
{
    "name": "smile-dental-agent-stub",
    "version": "1.0.0",
    "main": "agent.js",
    "scripts": {
        "start": "node agent.js"
    },
    "dependencies": {
        "ws": "^8.14.0"
    }
}
```

### Start Agent

```bash
cd agent-stub/

# Install dependencies (first time only)
npm install

# Start agent (default: number 101)
npm start

# Or with custom number/name
AGENT_NUMBER=101 AGENT_NAME="Smile Dental Receptionist" node agent.js

# Expected output:
# 🤖 Agent connecting to ws://localhost:3000/agent
#    Number: 101
#    Name: Smile Dental Receptionist
# ✅ Connected to backend
# ✅ Registered as 101 (Smile Dental Receptionist)
#    Waiting for incoming calls...
```

---

## 9.5 Complete System Startup

### Option A: Manual (3 Terminals)

```bash
# Terminal 1: Backend
cd backend && npm install && node server.js
# → Listening on ws://0.0.0.0:3000

# Terminal 2: Agent Stub
cd agent-stub && npm install && node agent.js
# → Registered as 101

# Terminal 3: Android
# Open android-app/ in Android Studio
# Set backend URL in res/values/config.xml
# Run on emulator or device
```

### Option B: Scripted (Single Command)

Create `start-all.sh` (Linux/Mac):
```bash
#!/bin/bash

echo "🦷 Starting Smile Dental System..."

# Start backend
echo "📡 Starting backend server..."
cd backend && npm install && node server.js &
BACKEND_PID=$!

# Wait for backend to start
sleep 2

# Start agent
echo "🤖 Starting agent stub..."
cd ../agent-stub && npm install && \
    AGENT_NUMBER=101 AGENT_NAME="Smile Dental Receptionist" node agent.js &
AGENT_PID=$!

echo ""
echo "✅ System started:"
echo "   Backend PID: $BACKEND_PID"
echo "   Agent PID: $AGENT_PID"
echo ""
echo "📱 Now open the Android app and dial 101"
echo ""
echo "To stop: kill PIDs or press Ctrl+C"

# Wait for interrupt
trap "kill $BACKEND_PID $AGENT_PID; exit" INT TERM
wait
```

Create `start-all.bat` (Windows):
```batch
@echo off
echo Starting Smile Dental System...

start "Backend" cmd /k "cd backend && npm install && node server.js"
timeout /t 2 /nobreak
start "Agent" cmd /k "cd agent-stub && npm install && set AGENT_NUMBER=101 && set AGENT_NAME=Smile Dental Receptionist && node agent.js"

echo.
echo System started. Open Android app and dial 101.
echo Close the terminal windows to stop.
```

---

## 9.6 Android Device Setup

### Configure Backend URL

Edit `res/values/config.xml`:
```xml
<resources>
    <!-- Replace with your machine's local IP -->
    <string name="backend_ws_url">ws://192.168.1.100:3000/call</string>
</resources>
```

### Find Your Local IP

```bash
# Windows
ipconfig | findstr IPv4

# Mac/Linux
ifconfig | grep "inet " | grep -v 127.0.0.1
```

### Install APK on Device

```bash
# Via ADB
adb install app/build/outputs/apk/debug/app-debug.apk

# Verify installation
adb shell pm list packages | grep smiledental
# Expected: package:com.smiledental.dialer
```

### Emulator Networking

If using Android Emulator, `localhost` refers to the emulator itself, not your host machine. Use `10.0.2.2` to reach the host:

```xml
<string name="backend_ws_url">ws://10.0.2.2:3000/call</string>
```

For physical devices on the same WiFi, use the host machine's actual IP:
```xml
<string name="backend_ws_url">ws://192.168.1.100:3000/call</string>
```

---

## 9.7 Deployment Checklist

### Pre-Deployment
- [ ] Backend server running and accessible on local network
- [ ] Agent(s) registered and showing as available
- [ ] Android APK built and signed (for release)
- [ ] Backend URL configured in Android app
- [ ] Firewall allows port 3000 (backend)

### Testing
- [ ] Dial 101 → connects to agent
- [ ] Audio loopback works (hear your own voice)
- [ ] Mute button works
- [ ] Hang-up button works
- [ ] Call ended screen shows duration
- [ ] Invalid number shows error
- [ ] Backend health check returns OK

### Production Considerations
- [ ] TLS enabled (wss://) for WebSocket
- [ ] Strong auth tokens generated
- [ ] Rate limiting configured
- [ ] Logging/monitoring enabled
- [ ] Auto-restart on crash (systemd/pm2)
- [ ] Backup phonebook.json

---

## 9.8 Production Deployment (systemd)

For Linux servers, run backend as a systemd service:

```ini
# /etc/systemd/system/smiledental-backend.service
[Unit]
Description=Smile Dental Backend Server
After=network.target

[Service]
Type=simple
User=smiledental
WorkingDirectory=/opt/smiledental/backend
ExecStart=/usr/bin/node server.js
Restart=on-failure
RestartSec=5
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable smiledental-backend
sudo systemctl start smiledental-backend

# Check status
sudo systemctl status smiledental-backend

# View logs
sudo journalctl -u smiledental-backend -f
```

---

*Next: [Page 10 — Testing Strategy](10-testing-strategy.md)*
