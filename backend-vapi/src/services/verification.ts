import crypto from "crypto";
import { sendVerificationEmail } from "./email";

// In-memory store — survives for the life of the container
// For multi-instance production: replace with Redis
interface VerificationRecord {
  email: string;
  token: string;
  verified: boolean;
  createdAt: number;
  verifiedAt?: number;
}

const store = new Map<string, VerificationRecord>(); // token → record
const byEmail = new Map<string, string>();             // email → token (latest)

const TTL_MS = 10 * 60 * 1000; // 10 minutes

// ─────────────────────────────────────────────────────────────────
// Create and send a verification link
// ─────────────────────────────────────────────────────────────────
export async function initiateVerification(email: string): Promise<"sent" | "error"> {
  // Invalidate any previous token for this email
  const prevToken = byEmail.get(email.toLowerCase());
  if (prevToken) store.delete(prevToken);

  const token = crypto.randomBytes(24).toString("hex");
  const record: VerificationRecord = {
    email: email.toLowerCase(),
    token,
    verified: false,
    createdAt: Date.now(),
  };

  store.set(token, record);
  byEmail.set(email.toLowerCase(), token);

  // Build the public URL
  const baseUrl = process.env.WEBHOOK_URL || `http://localhost:${process.env.PORT || 3000}`;
  const verifyUrl = `${baseUrl}/verify?token=${token}`;

  try {
    await sendVerificationEmail(email, verifyUrl);
    console.log(`🔐 Verification link sent to ${email}`);
    return "sent";
  } catch (err) {
    console.error("Failed to send verification email:", err);
    store.delete(token);
    byEmail.delete(email.toLowerCase());
    return "error";
  }
}

// ─────────────────────────────────────────────────────────────────
// Check status (called by Riley's tool)
// ─────────────────────────────────────────────────────────────────
export function checkVerification(email: string): "verified" | "pending" | "expired" | "not_found" {
  const token = byEmail.get(email.toLowerCase());
  if (!token) return "not_found";

  const record = store.get(token);
  if (!record) return "not_found";

  if (Date.now() - record.createdAt > TTL_MS) {
    store.delete(token);
    byEmail.delete(email.toLowerCase());
    return "expired";
  }

  return record.verified ? "verified" : "pending";
}

// ─────────────────────────────────────────────────────────────────
// Mark as verified (called when patient clicks the link)
// ─────────────────────────────────────────────────────────────────
export function markVerified(token: string): boolean {
  const record = store.get(token);
  if (!record) return false;

  if (Date.now() - record.createdAt > TTL_MS) {
    store.delete(token);
    byEmail.delete(record.email);
    return false;
  }

  record.verified = true;
  record.verifiedAt = Date.now();
  return true;
}

// Cleanup expired tokens periodically
setInterval(() => {
  const now = Date.now();
  for (const [token, record] of store) {
    if (now - record.createdAt > TTL_MS) {
      store.delete(token);
      byEmail.delete(record.email);
    }
  }
}, 5 * 60 * 1000);
