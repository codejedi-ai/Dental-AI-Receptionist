import { checkVerification } from "../services/verification";

export async function checkVerificationStatus(args: { email: string }): Promise<string> {
  const { email } = args;
  const status = checkVerification(email);

  switch (status) {
    case "verified":
      return "verified";
    case "pending":
      return "The link hasn't been clicked yet. Please ask the patient to check their inbox and click the verification link, then ask me to check again.";
    case "expired":
      return "The verification link has expired. Please send a new one using the send_verification_link tool.";
    case "not_found":
      return "No verification was initiated for that email. Please send a verification link first.";
  }
}
