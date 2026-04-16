import { Tool } from "../ToolDefinition.ts";
import { asObj, parseDateText } from "../lib/dentalHelpers.ts";

export const parse_date = new Tool(
  "parse_date",
  "Convert natural-language or relative dates to YYYY-MM-DD (Toronto).",
  "Use before check_availability when the user says 'tomorrow', weekday names, or Chinese equivalents. On failure, ask for YYYY-MM-DD.",
  {
    type: "object",
    properties: {
      date_text: {
        type: "string",
        description:
          "What the caller said (e.g. next Tuesday, tomorrow, 今天).",
      },
      dateText: {
        type: "string",
        description: "Alias for date_text.",
      },
    },
  },
  async (parameters) => {
    const a = asObj(parameters);
    const input = String(a.date_text ?? a.dateText ?? "");
    const parsed = parseDateText(input);
    if (!parsed) {
      return {
        result: `I couldn't parse '${input}'. Please use YYYY-MM-DD format.`,
      };
    }
    return { result: parsed };
  },
);
