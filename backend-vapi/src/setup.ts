/**
 * Setup script — creates the Vapi assistant and optionally a phone number.
 *
 * Run once: npm run setup
 */

import dotenv from "dotenv";
import { VapiClient } from "@vapi-ai/server-sdk";
import { ASSISTANT_CONFIG, TOOL_DEFINITIONS } from "./config/assistant";

dotenv.config();

const VAPI_API_KEY = process.env.VAPI_API_KEY;
if (!VAPI_API_KEY) {
  console.error("❌ Missing VAPI_API_KEY in .env");
  process.exit(1);
}

async function main() {
  const vapi = new VapiClient({ token: VAPI_API_KEY! });

  console.log("🦷 Setting up SmileDental Vapi assistant...\n");

  // ── Create Assistant ──
  try {
    const assistant = await vapi.assistants.create({
      name: ASSISTANT_CONFIG.name,
      firstMessage: ASSISTANT_CONFIG.firstMessage,
      model: {
        ...ASSISTANT_CONFIG.model,
        tools: TOOL_DEFINITIONS,
      } as any,
      voice: ASSISTANT_CONFIG.voice as any,
      transcriber: ASSISTANT_CONFIG.transcriber as any,
      silenceTimeoutSeconds: ASSISTANT_CONFIG.silenceTimeoutSeconds,
      maxDurationSeconds: ASSISTANT_CONFIG.maxDurationSeconds,
      endCallMessage: ASSISTANT_CONFIG.endCallMessage,
      serverUrl: process.env.WEBHOOK_URL
        ? `${process.env.WEBHOOK_URL}/webhook`
        : undefined,
      metadata: ASSISTANT_CONFIG.metadata,
    } as any);

    console.log("✅ Assistant created!");
    console.log(`   ID: ${assistant.id}`);
    console.log(`   Name: ${ASSISTANT_CONFIG.name}`);
    console.log("");

    // Save assistant ID for later use
    const fs = await import("fs");
    fs.writeFileSync(
      ".assistant-id",
      assistant.id || "",
      "utf-8"
    );
    console.log(`   Saved assistant ID to .assistant-id`);

    // ── Create Phone Number (optional) ──
    console.log("\n📞 To add a phone number:");
    console.log("   1. Go to https://dashboard.vapi.ai → Phone Numbers");
    console.log("   2. Create a new number and assign this assistant");
    console.log(`   3. Or use the API:`);
    console.log(`      await vapi.phoneNumbers.create({`);
    console.log(`        name: "SmileDental Main Line",`);
    console.log(`        assistantId: "${assistant.id}"`);
    console.log(`      });`);

    console.log("\n🌐 To test in browser:");
    console.log("   Open public/index.html");
    console.log(`   Use Public Key: ${process.env.VAPI_PUBLIC_KEY || "(set VAPI_PUBLIC_KEY in .env)"}`);
    console.log(`   Assistant ID: ${assistant.id}`);

    console.log("\n🔗 Webhook server:");
    console.log("   Run: npm run dev");
    console.log(`   Then set Server URL in Vapi dashboard to your webhook URL`);
  } catch (err) {
    console.error("❌ Failed to create assistant:", err);
    process.exit(1);
  }
}

main();
