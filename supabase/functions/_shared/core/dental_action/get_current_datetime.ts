import { Tool } from "../ToolDefinition.ts";

export const get_current_datetime = new Tool(
  "get_current_datetime",
  "Return the current Toronto date and time for scheduling context.",
  "Use when the caller asks for current date/time or when you need a precise now timestamp before booking decisions.",
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
      second: "2-digit",
      hour12: true,
    }).format(now);

    const parts = new Intl.DateTimeFormat("en-CA", {
      timeZone: "America/Toronto",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    }).formatToParts(now);

    const year = parts.find((p) => p.type === "year")?.value ?? "0000";
    const month = parts.find((p) => p.type === "month")?.value ?? "01";
    const day = parts.find((p) => p.type === "day")?.value ?? "01";
    const hour = parts.find((p) => p.type === "hour")?.value ?? "00";
    const minute = parts.find((p) => p.type === "minute")?.value ?? "00";
    const second = parts.find((p) => p.type === "second")?.value ?? "00";
    const isoLocal = `${year}-${month}-${day}T${hour}:${minute}:${second}`;

    return { result: `Current Toronto time is ${human}. ISO: ${isoLocal}` };
  },
);
