import { Tool } from "../ToolDefinition.ts";
import { getDentistsRaw } from "../lib/dentalHelpers.ts";

export const get_dentists = new Tool(
  "get_dentists",
  "List active dentists for scheduling.",
  "Call when the user asks who works at the clinic or before choosing a dentist for availability.",
  { type: "object", properties: {} },
  async (_parameters) => {
    try {
      const dentists = await getDentistsRaw();
      if (!dentists.length) {
        return {
          result: "I am having trouble retrieving the dentist list right now.",
        };
      }
      return {
        result: `Our dentists are: ${dentists.map((d) => d.name).join(", ")}.`,
      };
    } catch {
      return {
        result: "I am having trouble retrieving the dentist list right now.",
      };
    }
  },
);
