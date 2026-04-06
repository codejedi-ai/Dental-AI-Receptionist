/**
 * One-time Google OAuth2 setup.
 * Run:  npx tsx src/auth/google-setup.ts
 *
 * Prints a GOOGLE_REFRESH_TOKEN to add to your .env
 */
import { google } from "googleapis";
import * as readline from "readline";
import dotenv from "dotenv";
dotenv.config();

const SCOPES = ["https://www.googleapis.com/auth/calendar"];

async function main() {
  if (!process.env.GOOGLE_CLIENT_ID || !process.env.GOOGLE_CLIENT_SECRET) {
    console.error(`
❌ Missing GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET in .env

Steps to get them:
  1. Go to https://console.cloud.google.com
  2. Create a project (or select one)
  3. Enable "Google Calendar API"
  4. Go to APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client ID
  5. Application type: Desktop app
  6. Copy the Client ID and Client Secret into your .env:
       GOOGLE_CLIENT_ID=xxx
       GOOGLE_CLIENT_SECRET=xxx
  7. Re-run this script
    `);
    process.exit(1);
  }

  const oAuth2Client = new google.auth.OAuth2(
    process.env.GOOGLE_CLIENT_ID,
    process.env.GOOGLE_CLIENT_SECRET,
    "urn:ietf:wg:oauth:2.0:oob"
  );

  const authUrl = oAuth2Client.generateAuthUrl({
    access_type: "offline",
    scope: SCOPES,
    prompt: "consent",
  });

  console.log("\n📅 Google Calendar OAuth2 Setup\n");
  console.log("1. Open this URL in your browser:\n");
  console.log(`   ${authUrl}\n`);
  console.log("2. Sign in with darcy@cogni.inc");
  console.log("3. Allow calendar access");
  console.log("4. Copy the authorization code shown\n");

  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
  const code = await new Promise<string>((resolve) => {
    rl.question("Paste the authorization code here: ", (ans) => {
      rl.close();
      resolve(ans.trim());
    });
  });

  const { tokens } = await oAuth2Client.getToken(code);
  oAuth2Client.setCredentials(tokens);

  console.log("\n✅ Success! Add these to your .env:\n");
  console.log(`GOOGLE_REFRESH_TOKEN=${tokens.refresh_token}`);
  console.log(`GOOGLE_CALENDAR_ID=primary`);
  console.log("\nThen restart the container: docker compose restart");
}

main().catch(console.error);
