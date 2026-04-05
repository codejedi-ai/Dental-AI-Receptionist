# SmileDental — Vapi Voice Agent

A dental clinic voice AI agent built on [Vapi](https://vapi.ai).
Handles inbound/outbound calls for appointment booking, clinic info, and patient inquiries.

## Quick Start

### 1. Get Vapi API Keys
- Sign up at https://dashboard.vapi.ai
- Copy your **Private API Key** and **Public API Key**

### 2. Install dependencies
```bash
npm install
```

### 3. Configure environment
```bash
cp .env.example .env
# Edit .env with your Vapi API keys
```

### 4. Create the assistant + phone number
```bash
npm run setup
```

### 5. Start the webhook server
```bash
npm run dev
```

### 6. Test in browser
Open `public/index.html` in your browser, or call the assigned phone number.

## Architecture

```
backend-vapi/
├── src/
│   ├── setup.ts           # Creates Vapi assistant + phone number
│   ├── server.ts          # Express webhook server for tool calls
│   ├── tools/             # Custom tool definitions
│   │   ├── book-appointment.ts
│   │   ├── check-availability.ts
│   │   └── clinic-info.ts
│   └── config/
│       └── assistant.ts   # Assistant configuration (prompt, voice, model)
├── public/
│   └── index.html         # Web widget for browser-based calls
├── .env.example
├── package.json
└── tsconfig.json
```

## How It Works

1. **Patient calls** the clinic phone number (or clicks the web widget)
2. **Vapi answers** with a greeting using the configured voice
3. **Patient speaks** → Vapi transcribes (Deepgram) → LLM reasons (GPT-4o) → Vapi speaks (ElevenLabs)
4. **Tool calls**: When the patient wants to book, the LLM calls custom tools via webhook
5. **Webhook server** handles appointment logic and responds to Vapi

## Customization

- **Voice**: Change in `src/config/assistant.ts` (ElevenLabs, PlayHT, Rime, etc.)
- **LLM**: Swap model in assistant config (GPT-4o, Claude, Groq, etc.)
- **Transcriber**: Default Deepgram, can switch to OpenAI Whisper
- **Phone number**: Vapi provides numbers or bring your own via SIP/Twilio
