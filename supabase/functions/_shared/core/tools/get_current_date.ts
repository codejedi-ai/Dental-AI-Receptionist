import { Tool } from "../ToolDefinition.ts";

export const get_current_date = new Tool(
  "get_current_date",
  "Return the current date and time in Toronto timezone for scheduling context.",
  "Call before booking or when the user asks what day it is. Use the spoken result to anchor relative dates like 'tomorrow'.",
  { type: "object", properties: {} },
  async (_parameters) => {
    const now = new Date();
    const human = new Intl.DateTimeFormat("en-CA", {
      timeZone: "America/Toronto",
      weekday: "long",
      month: "long",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
    }).format(now);
    return { result: `Today is ${human} Toronto time.` };
  },
);
