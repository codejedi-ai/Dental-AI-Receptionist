import { DISPATCH_SUPPORTED_OPERATIONS } from "./lib/dispatchPayload.ts";
import { Tool } from "./ToolDefinition.ts";

/**
 * Registry for all Supabase-backed tools: handlers, docs, and router copy.
 */
export class SupabaseTool {
  private readonly byName = new Map<string, Tool>();

  constructor(tools: Tool[]) {
    for (const t of tools) {
      if (this.byName.has(t.name)) {
        throw new Error(`Duplicate tool name: ${t.name}`);
      }
      this.byName.set(t.name, t);
    }
  }

  /** Flat map for Vapi / legacy `defaultFunctions[name](args)` usage. */
  getHandlers(): Record<string, (parameters: unknown) => Promise<unknown>> {
    const out: Record<string, (parameters: unknown) => Promise<unknown>> = {};
    for (const [name, tool] of this.byName) {
      out[name] = (p) => tool.execute(p);
    }
    return out;
  }

  listOperationNames(): string[] {
    return [...this.byName.keys()].sort();
  }

  getTool(name: string): Tool | undefined {
    return this.byName.get(name);
  }

  /** Paragraph for `dispatch_dental_action` tool description in Vapi. */
  buildRouterDescription(): string {
    const ops = DISPATCH_SUPPORTED_OPERATIONS.join(", ");
    return (
      `Router for Supabase backend operations. Always pass payload (use {} if none). ` +
      `Optional requestId for idempotency. Operations: ${ops}. ` +
      "Payload fields map to each tool (see engineering docs); aliases phone→patientPhone, reason→service, provider→dentist are accepted."
    );
  }

  /** Compact per-operation guide intended for assistant system prompts. */
  buildPromptHandbook(): string {
    const lines: string[] = [];
    lines.push("=== TOOL HANDBOOK (AUTO-GENERATED) ===");
    lines.push("Use dispatch_dental_action with operation + payload (payload can be {}).");
    lines.push("Operations:");
    for (const name of this.listOperationNames()) {
      const t = this.byName.get(name)!;
      const required = Array.isArray(t.parametersSchema.required)
        ? t.parametersSchema.required.join(", ")
        : "";
      const requiredHint = required ? ` required: [${required}]` : "";
      lines.push(`- ${t.name}: ${t.description}${requiredHint}`);
    }
    lines.push("=== END TOOL HANDBOOK ===");
    return lines.join("\n");
  }

  /** Concatenated instruction manuals (markdown-ish) for debugging or prompts. */
  buildManualExport(): string {
    const lines: string[] = [];
    for (const name of this.listOperationNames()) {
      const t = this.byName.get(name)!;
      lines.push(
        `## ${t.name}`,
        "",
        t.description,
        "",
        "### Manual",
        t.manual,
        "",
        "### Parameters (JSON Schema)",
        JSON.stringify(t.parametersSchema, null, 2),
        "",
        "---",
        "",
      );
    }
    return lines.join("\n").trim();
  }
}
