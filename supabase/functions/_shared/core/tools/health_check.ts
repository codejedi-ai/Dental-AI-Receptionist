import { Tool } from "../ToolDefinition.ts";
import { getSupabaseConfig, restRequest } from "../lib/supabaseRest.ts";

export const health_check = new Tool(
  "health_check",
  "Verify Supabase connectivity and a minimal dentists table read.",
  "Run for diagnostics or startup checks. Does not modify data. Returns ok flag, base URL, and a probe row count.",
  { type: "object", properties: {} },
  async (_parameters) => {
    try {
      const cfg = getSupabaseConfig();
      const dentists = await restRequest<Array<{ id: number }>>(
        "GET",
        "dentists?select=id&limit=1",
      );
      return {
        result: "Harness and Supabase DB are healthy.",
        ok: true,
        service: "supabase-harness",
        baseUrl: cfg.baseUrl,
        dentistsProbeCount: dentists.length,
        timestamp: new Date().toISOString(),
      };
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        result: "Harness health check failed: " + message,
        ok: false,
        service: "supabase-harness",
        timestamp: new Date().toISOString(),
      };
    }
  },
);
