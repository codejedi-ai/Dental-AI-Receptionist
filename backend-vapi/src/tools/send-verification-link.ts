import { initiateVerification } from "../services/verification";

export async function sendVerificationLink(args: { email: string }): Promise<string> {
  const { email } = args;

  if (!email || !email.includes("@")) {
    return "That doesn't look like a valid email address. Could you spell it out again?";
  }

  const result = await initiateVerification(email);

  if (result === "sent") {
    return `I've sent a verification link to ${email}. Please check your inbox and click the link — it expires in 10 minutes. Once you've clicked it, let me know and I'll confirm the booking.`;
  } else {
    return "I wasn't able to send the verification email right now. Could you try a different email address, or would you like me to note your booking without email confirmation?";
  }
}
