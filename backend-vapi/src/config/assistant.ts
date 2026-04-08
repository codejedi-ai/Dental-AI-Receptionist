/**
 * SmileDental Vapi Assistant Configuration
 *
 * This defines the assistant's personality, voice, model, and tools.
 * Modify this to change how the dental receptionist behaves.
 */

export const SYSTEM_PROMPT = `You are Riley, the friendly AI receptionist at Smile Dental Clinic.

## Language
- Open every call with the bilingual greeting (already set as your first message).
- After the patient speaks, detect their language and respond entirely in that language for the rest of the call.
- You are fluent in English, Mandarin Chinese (普通话), and Cantonese (广东话).
- If the patient speaks Mandarin, respond in Mandarin. If Cantonese, respond in Cantonese.
- Do not switch languages unless the patient explicitly asks you to.
- When spelling back an email in Chinese, say each letter in Chinese: A=阿, B=波, C=西... then confirm "是这个邮箱吗？"

## Your Role
You answer phone calls and help patients with:
- Booking, rescheduling, or cancelling appointments
- Answering questions about services, hours, and location
- Providing pre-appointment instructions
- Routing urgent dental emergencies

## Clinic Information
- **Name**: Smile Dental Clinic
- **Address**: 123 Main Street, Newmarket, ON L3Y 4Z1
- **Phone**: (905) 555-0123
- **Hours**: Monday–Friday 8:00 AM – 6:00 PM, Saturday 9:00 AM – 2:00 PM, Closed Sunday
- **Emergency**: For after-hours dental emergencies, go to Southlake Regional Health Centre ER

## Services Offered
- General checkups & cleanings
- Teeth whitening
- Fillings & crowns
- Root canals
- Dental implants
- Invisalign & orthodontics
- Pediatric dentistry
- Emergency dental care

## Dentists
- Dr. Sarah Chen — General Dentistry, Cosmetic Dentistry
- Dr. Michael Park — Orthodontics, Invisalign
- Dr. Priya Sharma — Pediatric Dentistry, Root Canals

## Conversation Guidelines
- Be warm, patient, and reassuring (many patients are nervous about the dentist!)
- Keep responses concise — under 30 words when possible
- Always confirm details back to the patient before booking
- If you're unsure about something medical, say "Let me have the clinic follow up with you on that"
- For emergencies (severe pain, knocked-out tooth, heavy bleeding), advise going to ER immediately
- Speak naturally — use contractions, be conversational, not robotic
- When booking, collect: patient name, phone number, preferred date/time, reason for visit
- Offer the earliest available slot first

## Email Verification Flow (required before booking)
When a patient wants to book an appointment, you MUST verify their email before confirming:

1. **Ask for email**: "Could I get your email address to send you a confirmation?"
2. **Read it back letter-by-letter**: "Let me read that back: D as in Delta, A as in Alpha, R as in Romeo, C as in Charlie, Y as in Yankee, at G-mail dot com. Is that right?"
3. **Send the link**: Use the send_verification_link tool with their email.
4. **Wait**: Tell them "I've sent a verification link to your inbox. Please click it, then let me know when you're done."
5. **Check when ready**: When they say they've clicked it, use check_verification_status tool.
   - If it returns "verified" - proceed to book the appointment.
   - If pending - ask them to check again.
   - If expired - send a new link.
6. **Book only after verified**: Call book_appointment only after verification returns "verified".

## Important
- Never provide medical diagnosis
- Never discuss pricing over the phone — say "I'd be happy to get you a quote when you come in"
- If caller is rude or inappropriate, stay professional and offer to have the office manager call back`;

export const FIRST_MESSAGE =
  "Hi, thanks for calling Smile Dental! This is Riley. You can speak to me in English or Chinese — 您可以用英文或中文跟我交流。How can I help you today?";

export const ASSISTANT_CONFIG = {
  name: "SmileDental-Riley",
  firstMessage: FIRST_MESSAGE,

  // Model configuration
  model: {
    provider: "openai" as const,
    model: "gpt-4o",
    messages: [{ role: "system" as const, content: SYSTEM_PROMPT }],
    temperature: 0.7,
    maxTokens: 200,
  },

  // Voice configuration (ElevenLabs)
  voice: {
    provider: "11labs" as const,
    voiceId: "21m00Tcm4TlvDq8ikWAM", // "Rachel" — warm, professional female voice
    stability: 0.6,
    similarityBoost: 0.8,
  },

  // Transcriber configuration (Deepgram)
  transcriber: {
    provider: "deepgram" as const,
    model: "nova-2",
    language: "en",
  },

  // Speech settings
  silenceTimeoutSeconds: 20,
  maxDurationSeconds: 600, // 10 min max call
  endCallMessage: "Thanks for calling Smile Dental! Have a great day. Bye!",

  // Metadata
  metadata: {
    clinic: "smile-dental",
    version: "1.0",
  },
};

/**
 * Tool definitions for the assistant.
 * These are called via webhook when the LLM decides to use them.
 */
export const TOOL_DEFINITIONS = [
  {
    type: "function" as const,
    function: {
      name: "check_availability",
      description:
        "Check available appointment slots for a specific date and optionally a specific dentist. Call this when a patient wants to book an appointment.",
      parameters: {
        type: "object",
        properties: {
          date: {
            type: "string",
            description: "The date to check, in YYYY-MM-DD format",
          },
          dentist: {
            type: "string",
            description:
              "Optional: specific dentist name (Dr. Sarah Chen, Dr. Michael Park, Dr. Priya Sharma)",
            enum: ["Dr. Sarah Chen", "Dr. Michael Park", "Dr. Priya Sharma"],
          },
          service: {
            type: "string",
            description: "Type of appointment",
            enum: [
              "checkup",
              "cleaning",
              "whitening",
              "filling",
              "crown",
              "root-canal",
              "implant",
              "invisalign",
              "pediatric",
              "emergency",
            ],
          },
        },
        required: ["date"],
      },
    },
  },
  {
    type: "function" as const,
    function: {
      name: "book_appointment",
      description:
        "Book a confirmed appointment for a patient. Only call this after confirming all details with the patient.",
      parameters: {
        type: "object",
        properties: {
          patientName: {
            type: "string",
            description: "Full name of the patient",
          },
          patientPhone: {
            type: "string",
            description: "Patient phone number",
          },
          patientEmail: {
            type: "string",
            description: "Patient email address (required — must be verified before booking)",
          },
          date: {
            type: "string",
            description: "Appointment date in YYYY-MM-DD format",
          },
          time: {
            type: "string",
            description: "Appointment time in HH:MM format (24h)",
          },
          dentist: {
            type: "string",
            description: "Dentist name",
            enum: ["Dr. Sarah Chen", "Dr. Michael Park", "Dr. Priya Sharma"],
          },
          service: {
            type: "string",
            description: "Type of appointment",
          },
          notes: {
            type: "string",
            description: "Any special notes or concerns from the patient",
          },
        },
        required: ["patientName", "patientPhone", "date", "time", "dentist", "service"],
      },
    },
  },
  {
    type: "function" as const,
    function: {
      name: "cancel_appointment",
      description: "Cancel an existing appointment by patient name and date.",
      parameters: {
        type: "object",
        properties: {
          patientName: {
            type: "string",
            description: "Full name of the patient",
          },
          patientPhone: {
            type: "string",
            description: "Patient phone number for verification",
          },
          date: {
            type: "string",
            description: "The date of the appointment to cancel (YYYY-MM-DD)",
          },
        },
        required: ["patientName", "date"],
      },
    },
  },
  {
    type: "function" as const,
    function: {
      name: "get_clinic_info",
      description:
        "Get information about the clinic such as hours, location, insurance accepted, emergency procedures, cancellation policy, new patient instructions, or payment options.",
      parameters: {
        type: "object",
        properties: {
          topic: {
            type: "string",
            description: "The topic of information needed (hours, location, insurance, emergency, cancellation, newpatient, payment)",
            enum: ["hours", "location", "insurance", "emergency", "cancellation", "newpatient", "payment"],
          },
        },
        required: ["topic"],
      },
    },
  },
  {
    type: "function" as const,
    function: {
      name: "send_verification_link",
      description:
        "Send an email verification link to the patient's email address. Call this after spelling the email back to the patient and they confirm it's correct. The patient must click the link before you can book.",
      parameters: {
        type: "object",
        properties: {
          email: {
            type: "string",
            description: "The patient's email address",
          },
        },
        required: ["email"],
      },
    },
  },
  {
    type: "function" as const,
    function: {
      name: "check_verification_status",
      description:
        "Check whether the patient has clicked the email verification link. Returns 'verified' if confirmed. Call this when the patient says they've clicked the link.",
      parameters: {
        type: "object",
        properties: {
          email: {
            type: "string",
            description: "The patient's email address",
          },
        },
        required: ["email"],
      },
    },
  },
  {
    type: "function" as const,
    function: {
      name: "get_current_date",
      description:
        "Get the current date and time in the clinic's timezone (America/Toronto). Use this when the patient asks about dates, days of the week, or when you need to know what today's date is.",
      parameters: {
        type: "object",
        properties: {},
        required: [],
      },
    },
  },
];
