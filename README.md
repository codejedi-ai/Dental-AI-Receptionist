# 🦷 SmileDental — AI Dental Receptionist

An AI-powered dental clinic system with a voice receptionist named **Lisa**. Patients can book appointments via a web app or by talking to Lisa through the Vapi voice widget or phone.

## Architecture

```
SmileDental/
├── app/                    # Vite + React clinic/admin frontend
│   ├── src/pages/          # Patient and admin pages
│   └── src/components/     # UI components + Voice widget
├── backend/                # Unified backend (Vapi tools + clinic API)
│   ├── src/routes/         # REST API (appointments, patients)
│   └── src/tools/          # Vapi tool handlers
└── engineering-notebook/   # System design documentation
```

## Quick Start (Web App)

```bash
cd app
npm install
npm run dev
# Open http://localhost:5173
```

## Features

- 🏠 **Patient Website** — Modern clinic homepage, services, dentist profiles
- 📅 **Online Booking** — Step-by-step appointment booking with real-time availability
- 👩‍💼 **Admin Dashboard** — Manage appointments, patients, view stats
- 🎙️ **Voice AI Widget** — Talk to Lisa (Vapi) to book appointments by voice
- ☎️ **Phone Support** — Assign a phone number via Vapi for real phone calls
- 🗄️ **SQLite Database** — Zero-config, auto-creates on first run

## Tech Stack

- **Vite + React** + **TypeScript** + **Tailwind CSS**
- **SQLite** via better-sqlite3
- **JSON object store** in `backend/data/appointments/` for tool-side persisted artifacts
- **Vapi** for voice AI (STT → LLM → TTS)
- **GPT-4o** for conversation, **Deepgram** for transcription, **ElevenLabs** for voice

## Clinic Info

- **Name:** Smile Dental Clinic
- **Location:** 123 Main Street, Newmarket, ON
- **Hours:** Mon–Fri 8am–6pm, Sat 9am–2pm
- **Dentists:** Dr. Sarah Chen, Dr. Michael Park, Dr. Priya Sharma

## License

MIT
