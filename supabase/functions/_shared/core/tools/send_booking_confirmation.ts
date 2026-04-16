import { Tool } from "../ToolDefinition.ts";

export const send_booking_confirmation = new Tool(
  "send_booking_confirmation",
  "Acknowledge that a confirmation message will be sent (SMS/email) after booking.",
  "Call immediately after a successful book_appointment when workflow should mention confirmation delivery.",
  { type: "object", properties: {} },
  async (_parameters) => ({
    result:
      "Your appointment has been confirmed. You will receive a text or email confirmation with the details shortly.",
  }),
);
