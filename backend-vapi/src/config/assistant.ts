/**
 * SmileDental Vapi Assistant Configuration
 *
 * This defines the assistant's personality, voice, model, and tools.
 * Modify this to change how the dental receptionist behaves.
 */

export const SYSTEM_PROMPT = `You are Lisa, the friendly AI receptionist at Smile Dental Clinic.

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

## Important
- Never provide medical diagnosis
- Never discuss pricing over the phone — say "I'd be happy to get you a quote when you come in"
- If caller is rude or inappropriate, stay professional and offer to have the office manager call back`;

export const FIRST_MESSAGE =
  "Hi! Thanks for calling Smile Dental Clinic. This is Lisa. How can I help you today?";

export const ASSISTANT_CONFIG = {
  name: "SmileDental-Lisa",
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
              "Optional: specific dentist name (Dr. Chen, Dr. Park, Dr. Sharma)",
            enum: ["Dr. Chen", "Dr. Park", "Dr. Sharma"],
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
            enum: ["Dr. Chen", "Dr. Park", "Dr. Sharma"],
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
];
