/**
 * SmileDental Vapi Webhook Server
 *
 * Handles tool calls from Vapi when the assistant needs to
 * check availability, book appointments, etc.
 *
 * Vapi sends POST requests to your webhook URL with tool call payloads.
 */

import express from "express";
import dotenv from "dotenv";
import { checkAvailability } from "./tools/check-availability";
import { bookAppointment } from "./tools/book-appointment";
import { cancelAppointment } from "./tools/cancel-appointment";
import { getClinicInfo } from "./tools/clinic-info";

dotenv.config();

const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3000;

// ─── Health Check ──────────────────────────────────────────────
app.get("/health", (_req, res) => {
  res.json({ status: "ok", service: "smiledental-vapi" });
});

// ─── Vapi Webhook (Server URL) ─────────────────────────────────
// Vapi sends various message types to your server URL.
// See: https://docs.vapi.ai/server-url
app.post("/webhook", (req, res) => {
  const { message } = req.body;

  console.log(`📞 Webhook received: ${message?.type || "unknown"}`);

  switch (message?.type) {
    // ── Tool Calls ──
    case "tool-calls": {
      const toolCalls = message.toolCalls || message.toolCallList || [];
      const results = toolCalls.map((toolCall: any) => {
        const { name, arguments: args } = toolCall.function;
        console.log(`🔧 Tool call: ${name}`, args);

        let result: string;
        try {
          switch (name) {
            case "check_availability":
              result = checkAvailability(args);
              break;
            case "book_appointment":
              result = bookAppointment(args);
              break;
            case "cancel_appointment":
              result = cancelAppointment(args);
              break;
            case "get_clinic_info":
              result = getClinicInfo(args);
              break;
            default:
              result = `Unknown tool: ${name}`;
          }
        } catch (err) {
          console.error(`❌ Tool error (${name}):`, err);
          result = "Sorry, I had trouble processing that. Let me try again.";
        }

        return {
          toolCallId: toolCall.id,
          result,
        };
      });

      return res.json({ results });
    }

    // ── Status Updates ──
    case "status-update": {
      const { status } = message;
      console.log(`📊 Call status: ${status.status}`);

      if (status.status === "ended") {
        console.log("📱 Call ended.", {
          duration: status.endedReason,
          messages: status.messages?.length || 0,
        });

        // In production: save call transcript, update CRM, send follow-up SMS, etc.
      }
      return res.json({});
    }

    // ── Conversation Update ──
    case "conversation-update": {
      // Real-time transcript updates — useful for live dashboards
      return res.json({});
    }

    // ── End of Call Report ──
    case "end-of-call-report": {
      console.log("📋 End of call report:", {
        summary: message.summary,
        duration: message.durationSeconds,
        cost: message.cost,
      });

      // In production: save report to database, trigger follow-up workflows
      return res.json({});
    }

    // ── Speech Update ──
    case "speech-update": {
      return res.json({});
    }

    // ── Hang ──
    case "hang": {
      console.log("⏳ Hang detected — assistant is processing");
      return res.json({});
    }

    default: {
      console.log(`ℹ️ Unhandled message type: ${message?.type}`);
      return res.json({});
    }
  }
});

// ─── Start Server ──────────────────────────────────────────────
app.listen(PORT, () => {
  console.log(`\n🦷 SmileDental Vapi webhook server running on port ${PORT}`);
  console.log(`   Health: http://localhost:${PORT}/health`);
  console.log(`   Webhook: http://localhost:${PORT}/webhook`);
  console.log(`\n   Set your Vapi assistant's Server URL to:`);
  console.log(`   ${process.env.WEBHOOK_URL || `http://localhost:${PORT}`}/webhook\n`);
});
