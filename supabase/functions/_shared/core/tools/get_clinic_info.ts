import { Tool } from "../ToolDefinition.ts";
import { CLINIC_INFO } from "../lib/dentalHelpers.ts";

export const get_clinic_info = new Tool(
  "get_clinic_info",
  "Return static clinic information: hours, services, location, insurance, emergency.",
  "Always use this instead of guessing clinic facts. Topic selects which paragraph; omit or use general for a default overview.",
  {
    type: "object",
    properties: {
      topic: {
        type: "string",
        description:
          "One of: general, hours, services, location, insurance, emergency.",
      },
    },
  },
  async (parameters) => {
    const a = parameters as Record<string, unknown>;
    const topic = String(a?.topic ?? "general").toLowerCase();
    return { result: CLINIC_INFO[topic] ?? CLINIC_INFO.general };
  },
);
