# Page 13 — Module 1: Language Detection

---

## 13.1 Overview

Module 1 handles **language detection and initial response generation** for the AI Dental Receptionist (Lisa). When a patient first speaks to the system, this module detects whether they are speaking English or Chinese, sets the conversation language accordingly, and generates a warm, polite first response — all within ≤2 sentences.

This module is the entry point for every conversation. Its output (`lang_code` + `first_response`) is passed downstream to **Module 2** for intent classification and task execution.

---

## 13.2 Flow Diagram

```
                    ┌──────────────────────────┐
                    │  Module 1: Language       │
                    │                           │
                    │  ┌─────────────────────┐  │
                    │  │ Input:              │  │
                    │  │ user_first_sentence │  │
                    │  └──────────┬──────────┘  │
                    │             │              │
                    │             ▼              │
                    │  ┌─────────────────────┐  │
                    │  │   Detect language    │  │
                    │  └────────────────────┘  │
                    │             │              │
                    │      ┌──────┼──────┐       │
                    │      │      │      │       │
                    │      ▼      ▼      ▼       │
                    │  Contains English   Uncertain│
                    │  Chinese  only      / mixed │
                    │      │      │      │       │
                    │      ▼      ▼      ▼       │
                    │   Set     Set     Set      │
                    │  lang=   lang=   lang='en' │
                    │  'zh'    'en'    + log      │
                    │      │      │   uncertain   │
                    │      │      │      │       │
                    │      ▼      └──┬───┘       │
                    │   Reply in     │            │
                    │   zh, ≤2 sent  │            │
                    │   neutral/warm │            │
                    │   polite/prof  │            │
                    │      │         │            │
                    │      ▼         ▼            │
                    │   Reply in en, ≤2 sentences │
                    │   neutral/warm/polite/prof  │
                    │      │         │            │
                    │      └────┬────┘            │
                    │           │                  │
                    │           ▼                  │
                    │  ┌─────────────────────┐    │
                    │  │ Output:             │    │
                    │  │ lang_code +         │    │
                    │  │ first_response      │    │
                    │  └──────────┬──────────┘    │
                    │             │                │
                    └─────────────┼────────────────┘
                                  │
                     lang_code + first_response
                                  │
                                  ▼
                          ┌──────────────┐
                          │  To Module 2  │
                          │  (Intent      │
                          │   Classifier) │
                          └──────────────┘
```

---

## 13.3 Language Detection Logic

### Detection Rules

| Condition | Detection Rule | `lang` Value |
|-----------|---------------|--------------|
| Input contains **any Chinese characters** (Unicode range `\u4e00`–`\u9fff`) | `[\u4e00-\u9fff]` regex match | `'zh'` |
| Input contains **only ASCII/Latin characters** (no CJK) | No Chinese chars detected | `'en'` |
| Input is **uncertain or mixed** (e.g., pinyin + English, ambiguous) | Heuristic unclear | `'en'` + log warning |

### Regex Pattern

```javascript
const CHINESE_CHAR_PATTERN = /[\u4e00-\u9fff]/;

function detectLanguage(sentence) {
  if (CHINESE_CHAR_PATTERN.test(sentence)) {
    return 'zh';
  }
  // Default to English for ASCII-only or uncertain input
  return 'en';
}
```

### Edge Cases

| Input | Detected Language | Reason |
|-------|-------------------|--------|
| `"你好，我想预约"` | `zh` | Contains Chinese characters |
| `"I'd like to book an appointment"` | `en` | English only |
| `"Hello 你好"` | `zh` | Contains at least one Chinese character |
| `"Ni hao"` | `en` | Pinyin is Latin characters, no Hanzi |
| `"..."` or empty | `en` | No Chinese chars, defaults to English |
| `"I need 牙医"` | `zh` | Mixed, but contains Chinese character |

---

## 13.4 Response Generation

### Tone Guidelines

All first responses must be:

| Attribute | Description |
|-----------|-------------|
| **Neutral** | Objective, non-presumptuous |
| **Warm** | Friendly and welcoming |
| **Polite** | Respectful, uses courteous phrasing |
| **Professional** | Clinic-appropriate formality |

### Response Length

- **Maximum 2 sentences** — keeps the response concise and leaves room for the patient to continue.
- Avoid information overload on first turn.

### Template Responses

#### English (`lang = 'en'`)

| Scenario | Response |
|----------|----------|
| General greeting | `"Hello! Welcome to Smile Dental Clinic. How can I help you today?"` |
| Patient states intent directly | `"Thank you for calling Smile Dental. I'd be happy to help you with that."` |
| Unclear input | `"Hello! This is Lisa from Smile Dental. Could you tell me how I can assist you?"` |

#### Chinese (`lang = 'zh'`)

| Scenario | Response |
|----------|----------|
| General greeting | `"您好！欢迎来到微笑牙科。请问今天有什么可以帮您的？"` |
| Patient states intent directly | `"感谢您致电微笑牙科。我很乐意为您安排。"` |
| Unclear input | `"您好！我是微笑牙科的Lisa。请问您需要什么帮助呢？"` |

### Dynamic Response Generation (LLM-based)

If using an LLM (e.g., GPT-4o) for first-response generation instead of static templates:

**System Prompt (English):**
```
You are Lisa, the AI receptionist at Smile Dental Clinic.
Generate a warm, polite, professional first response in {lang}.
Keep it to at most 2 sentences.
Do not ask multiple questions. Do not provide medical advice.
Clinic name: Smile Dental Clinic
Location: 123 Main Street, Newmarket, ON
```

**System Prompt (Chinese):**
```
你是微笑牙科的AI接待员Lisa。
请用{lang}生成一段热情、礼貌、专业的首次回复。
不超过2句话。
不要问多个问题，不要提供医疗建议。
诊所名称：微笑牙科
地址：安大略省纽马克特市主街123号
```

---

## 13.5 Output Format

The module produces a structured output consumed by Module 2:

```json
{
  "module": "language_detection",
  "lang_code": "zh",
  "first_response": "您好！欢迎来到微笑牙科。请问今天有什么可以帮您的？",
  "confidence": "high",
  "metadata": {
    "input_sentence": "你好，我想预约",
    "detected_chinese": true,
    "detection_method": "regex"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `module` | `string` | Always `"language_detection"` |
| `lang_code` | `"en" \| "zh"` | Detected language code |
| `first_response` | `string` | Generated reply in detected language |
| `confidence` | `"high" \| "uncertain"` | Detection confidence level |
| `metadata.input_sentence` | `string` | Original user input |
| `metadata.detected_chinese` | `boolean` | Whether Chinese characters were found |
| `metadata.detection_method` | `string` | How detection was performed |

---

## 13.6 Integration with Vapi

Since this system uses **Vapi** for voice AI, Module 1 integrates as follows:

### Vapi Assistant Configuration

```json
{
  "assistant": {
    "name": "Lisa - Smile Dental Receptionist",
    "firstMessageMode": "assistant-speaks-first",
    "firstMessage": "Hello! Welcome to Smile Dental Clinic. How can I help you today?",
    "model": {
      "provider": "openai",
      "model": "gpt-4o",
      "messages": [
        {
          "role": "system",
          "content": "You are Lisa, an AI dental receptionist. Detect the patient's language from their first utterance. If they speak Chinese, respond in Chinese. Otherwise, respond in English. Keep responses concise (≤2 sentences for greetings). Be warm, polite, and professional."
        }
      ]
    },
    "voice": {
      "provider": "11labs",
      "voiceId": "lisa-voice-id",
      "stability": 0.5,
      "similarityBoost": 0.75
    }
  }
}
```

### Pre-Processing Hook (Backend)

If language detection is performed server-side before Vapi processes the utterance:

```javascript
// backend/src/tools/language-detection.js
const { CHINESE_CHAR_PATTERN } = require('../constants');

async function preProcessFirstUtterance(text) {
  const langCode = CHINESE_CHAR_PATTERN.test(text) ? 'zh' : 'en';

  if (langCode === 'en' && !/^[a-zA-Z\s.,!?'"-]*$/.test(text)) {
    console.warn(`[LanguageDetection] Uncertain input: "${text}", defaulting to 'en'`);
  }

  return { langCode, text };
}

module.exports = { preProcessFirstUtterance };
```

---

## 13.7 Logging & Monitoring

### Log Events

| Event | Level | Log Message |
|-------|-------|-------------|
| Chinese detected | `INFO` | `[LangDetect] Detected Chinese: "{input}" → lang=zh` |
| English detected | `INFO` | `[LangDetect] Detected English: "{input}" → lang=en` |
| Uncertain/mixed | `WARN` | `[LangDetect] Uncertain input: "{input}" → lang=en (logged for review)` |
| Empty input | `WARN` | `[LangDetect] Empty or whitespace-only input` |

### Metrics (Optional)

| Metric | Description |
|--------|-------------|
| `lang_detection.total` | Total detections |
| `lang_detection.by_language` | Count by detected language (`zh`, `en`) |
| `lang_detection.uncertain` | Count of uncertain detections |
| `lang_detection.latency_ms` | Detection time (should be <10ms) |

---

## 13.8 Error Handling

| Error Condition | Handling |
|-----------------|----------|
| Input is `null` or `undefined` | Default to `lang='en'`, log error, use generic greeting |
| Input is empty string/whitespace | Default to `lang='en'`, prompt patient to speak |
| Input contains only punctuation | Default to `lang='en'`, ask for clarification |
| Detection timeout (async) | Not applicable — detection is synchronous regex, should not timeout |

---

## 13.9 Test Cases

| # | Input | Expected `lang` | Expected Response Lang |
|---|-------|-----------------|------------------------|
| 1 | `"你好"` | `zh` | Chinese |
| 2 | `"Hello"` | `en` | English |
| 3 | `"I want 预约"` | `zh` | Chinese |
| 4 | `"Ni hao ma"` | `en` | English (pinyin only) |
| 5 | `"..."` | `en` | English (prompt for input) |
| 6 | `"Can I book an appointment for 明天?"` | `zh` | Chinese |
| 7 | `""` (empty) | `en` | English (prompt for input) |
| 8 | `"你好！Hello!"` | `zh` | Chinese |
| 9 | `"Is this 牙科诊所?"` | `zh` | Chinese |
| 10 | `"Good morning"` | `en` | English |

---

## 13.10 Implementation Checklist

- [ ] Create `backend/src/tools/language-detection.js` module
- [ ] Implement regex-based Chinese character detection
- [ ] Add response templates for English and Chinese
- [ ] Integrate with Vapi assistant system prompt
- [ ] Add logging for all detection events
- [ ] Write unit tests for all edge cases
- [ ] Update Vapi assistant config with bilingual system prompt
- [ ] Add metrics collection for detection accuracy
- [ ] Document handoff protocol to Module 2

---

*Previous: [Page 12 — Future Roadmap](12-future-roadmap.md)*
*Next: [Page 14 — Module 2: Intent Classification](14-module-2-intent-classification.md)*
