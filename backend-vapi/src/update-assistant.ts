/**
 * Update the existing Riley assistant on Vapi with the latest
 * system prompt, tools, and webhook URL.
 *
 * Run: npx tsx src/update-assistant.ts
 */

import dotenv from "dotenv";
import { VapiClient } from "@vapi-ai/server-sdk";
import { SYSTEM_PROMPT, TOOL_DEFINITIONS } from "./config/assistant";

dotenv.config();

const VAPI_API_KEY    = process.env.VAPI_API_KEY!;
const ASSISTANT_ID    = process.env.VAPI_ASSISTANT_ID || "450435e9-4562-4ddd-8429-54584d3285a7";
const WEBHOOK_URL     = process.env.WEBHOOK_URL || "http://localhost:3000";

if (!VAPI_API_KEY) {
  console.error("❌ Missing VAPI_API_KEY in .env");
  process.exit(1);
}

async function main() {
  const vapi = new VapiClient({ token: VAPI_API_KEY });

  console.log(`🦷 Updating Riley (${ASSISTANT_ID})…\n`);

  try {
    const updated = await (vapi.assistants as any).update(ASSISTANT_ID, {
      model: {
        provider: "openai",
        model: "gpt-4o",
        messages: [{ role: "system", content: SYSTEM_PROMPT }],
        temperature: 0.7,
        maxTokens: 200,
        tools: TOOL_DEFINITIONS,
      },
      transcriber: {
        provider: "deepgram",
        model: "nova-2",
        language: "multi",   // enables multi-language detection
      },
      serverUrl: `${WEBHOOK_URL}/webhook`,
    });

    console.log("✅ Riley updated successfully!");
    console.log(`   ID:          ${updated.id}`);
    console.log(`   Server URL:  ${WEBHOOK_URL}/webhook`);
    console.log(`   Language:    multi (EN + ZH + more)`);
    console.log(`   Tools:       ${TOOL_DEFINITIONS.length} tools registered`);
  } catch (err: any) {
    console.error("❌ Update failed:", err?.message || err);
    process.exit(1);
  }
}

main();
