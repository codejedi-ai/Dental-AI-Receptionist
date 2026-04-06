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
import { sendVerificationLink } from "./tools/send-verification-link";
import { checkVerificationStatus } from "./tools/check-verification-status";
import { getCurrentDate } from "./tools/get-current-date";
import { markVerified } from "./services/verification";

dotenv.config();

const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3000;

// ─── Health Check ──────────────────────────────────────────────
app.get("/health", (_req, res) => {
  res.json({ status: "ok", service: "smiledental-vapi" });
});

// ─── Email Verification ────────────────────────────────────────
app.get("/verify", (req, res) => {
  const token = req.query.token as string | undefined;
  if (!token) {
    return res.status(400).send(verifyPage("Invalid link", "This verification link is invalid.", false));
  }
  const ok = markVerified(token);
  if (ok) {
    console.log(`✅ Email verified via token: ${token.slice(0, 8)}…`);
    return res.send(verifyPage("Email Verified!", "Your email has been verified. You're all set — your appointment will be confirmed shortly.", true));
  } else {
    return res.status(410).send(verifyPage("Link Expired", "This verification link has expired or was already used. Please call us if you need to rebook.", false));
  }
});

function verifyPage(title: string, body: string, success: boolean): string {
  const color = success ? "#1a6fc4" : "#c0392b";
  const icon  = success ? "✅" : "⚠️";
  return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>${title} — Smile Dental</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{font-family:sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#f4f6f9}
.card{background:#fff;border-radius:12px;padding:40px 32px;max-width:420px;text-align:center;box-shadow:0 4px 20px rgba(0,0,0,.1)}
h1{color:${color};margin-top:8px}p{color:#555;line-height:1.6}.icon{font-size:48px}</style></head>
<body><div class="card"><div class="icon">${icon}</div><h1>${title}</h1><p>${body}</p>
<p style="margin-top:32px;font-size:13px;color:#999">Smile Dental Clinic · (905) 555-0123</p></div></body></html>`;
}

// ─── Vapi Webhook (Server URL) ─────────────────────────────────
// Vapi sends various message types to your server URL.
// See: https://docs.vapi.ai/server-url
app.post("/webhook", async (req, res) => {
  const { message } = req.body;

  console.log(`📞 Webhook received: ${message?.type || "unknown"}`);

  switch (message?.type) {
    // ── Tool Calls ──
    case "tool-calls": {
      const toolCalls = message.toolCalls || message.toolCallList || [];
      const results = await Promise.all(toolCalls.map(async (toolCall: any) => {
        const { name, arguments: args } = toolCall.function;
        console.log(`🔧 Tool call: ${name}`, args);

        let result: string;
        try {
          switch (name) {
            case "check_availability":
              result = await checkAvailability(args);
              break;
            case "book_appointment":
              result = await bookAppointment(args);
              break;
            case "cancel_appointment":
              result = await cancelAppointment(args);
              break;
            case "get_clinic_info":
              result = getClinicInfo(args);
              break;
            case "send_verification_link":
              result = await sendVerificationLink(args);
              break;
            case "check_verification_status":
              result = await checkVerificationStatus(args);
              break;
            case "get_current_date":
              result = getCurrentDate();
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
      }));

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
