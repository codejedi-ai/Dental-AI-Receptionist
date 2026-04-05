/**
 * Test script — triggers an outbound test call via Vapi.
 *
 * Usage: npm run test-call
 *
 * Make sure to set TEST_PHONE_NUMBER in your .env file.
 */

import dotenv from "dotenv";
import { VapiClient } from "@vapi-ai/server-sdk";
import fs from "fs";

dotenv.config();

const VAPI_API_KEY = process.env.VAPI_API_KEY;
const TEST_PHONE = process.env.TEST_PHONE_NUMBER;

if (!VAPI_API_KEY) {
  console.error("❌ Missing VAPI_API_KEY in .env");
  process.exit(1);
}

async function main() {
  const vapi = new VapiClient({ token: VAPI_API_KEY! });

  // Read assistant ID from setup
  let assistantId: string;
  try {
    assistantId = fs.readFileSync(".assistant-id", "utf-8").trim();
  } catch {
    console.error("❌ No .assistant-id file found. Run 'npm run setup' first.");
    process.exit(1);
  }

  if (!TEST_PHONE) {
    console.log("ℹ️  No TEST_PHONE_NUMBER in .env — listing assistant info instead.\n");

    const assistant = await vapi.assistants.get(assistantId);
    console.log("🦷 SmileDental Assistant:");
    console.log(`   ID: ${assistant.id}`);
    console.log(`   Name: ${(assistant as any).name}`);
    console.log(`   Created: ${(assistant as any).createdAt}`);
    console.log("\n   To make a test call, add TEST_PHONE_NUMBER=+1XXXXXXXXXX to .env");
    return;
  }

  console.log(`📞 Calling ${TEST_PHONE} with SmileDental assistant...\n`);

  try {
    const call = await vapi.calls.create({
      assistantId,
      customer: { number: TEST_PHONE },
    });

    console.log("✅ Call initiated!");
    console.log(`   Call ID: ${call.id}`);
    console.log(`   Status: ${call.status}`);
    console.log("\n   The phone should ring shortly. Pick up to test the assistant!");
  } catch (err) {
    console.error("❌ Failed to create call:", err);
  }
}

main();
