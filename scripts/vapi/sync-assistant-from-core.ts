import { DISPATCH_DENTAL_ACTION_PARAMETER_SCHEMA } from "../../supabase/functions/_shared/core/lib/dispatchPayload.ts";
import { supabaseTool } from "../../supabase/functions/_shared/function-call/dentalTools.ts";

const ASSISTANT_PATH = new URL("../../vapi/riley-assistant.json", import.meta.url);
const START_MARKER = "=== TOOL HANDBOOK (AUTO-GENERATED) ===";
const END_MARKER = "=== END TOOL HANDBOOK ===";
const TOOL_FIRST_MARKER = "=== TOOL-FIRST POLICY (AUTO-GENERATED) ===";
const TOOL_FIRST_POLICY = [
  TOOL_FIRST_MARKER,
  "Every dental_action is an action the AI can do with the tool.",
  "You MUST call dispatch_dental_action before any final answer for clinic data.",
  "If a request needs clinic facts, dentist list, date parsing, availability, booking, cancellation, or confirmation, call the tool first.",
  "For operations with no parameters, send payload: {} explicitly.",
  "Do NOT invent clinic data or scheduling data from memory.",
  "If tool output is missing or errors, explain the failure briefly and ask one follow-up question.",
  "=== END TOOL-FIRST POLICY ===",
].join("\n");

function stripOldHandbook(systemPrompt: string): string {
  const start = systemPrompt.indexOf(START_MARKER);
  if (start === -1) return systemPrompt.trim();
  const end = systemPrompt.indexOf(END_MARKER, start);
  if (end === -1) return systemPrompt.slice(0, start).trim();
  return systemPrompt.slice(0, start).trim();
}

const raw = await Deno.readTextFile(ASSISTANT_PATH);
const assistant = JSON.parse(raw);

const model = assistant.model ?? (assistant.model = {});
const tools = Array.isArray(model.tools) ? model.tools : [];
const fnTool = tools.find((t: unknown) =>
  typeof t === "object" &&
  t !== null &&
  (t as Record<string, unknown>).type === "function" &&
  typeof (t as Record<string, unknown>).function === "object" &&
  ((t as Record<string, unknown>).function as Record<string, unknown>).name ===
    "dispatch_dental_action"
) as Record<string, unknown> | undefined;

if (!fnTool || typeof fnTool.function !== "object" || fnTool.function === null) {
  throw new Error("Could not find dispatch_dental_action function tool in riley-assistant.json");
}

const fn = fnTool.function as Record<string, unknown>;
fn.description = supabaseTool.buildRouterDescription();
fn.parameters = DISPATCH_DENTAL_ACTION_PARAMETER_SCHEMA;

const messages = Array.isArray(model.messages) ? model.messages : [];
const systemMsg = messages.find((m: unknown) =>
  typeof m === "object" &&
  m !== null &&
  (m as Record<string, unknown>).role === "system" &&
  typeof (m as Record<string, unknown>).content === "string"
) as Record<string, unknown> | undefined;

if (!systemMsg) {
  throw new Error("No system message found in model.messages");
}

const basePrompt = stripOldHandbook(String(systemMsg.content ?? ""));
const handbook = supabaseTool.buildPromptHandbook();
const baseWithoutPolicy = basePrompt.includes(TOOL_FIRST_MARKER)
  ? basePrompt.slice(0, basePrompt.indexOf(TOOL_FIRST_MARKER)).trim()
  : basePrompt.trim();
systemMsg.content = `${baseWithoutPolicy}\n\n${TOOL_FIRST_POLICY}\n\n${handbook}`;

await Deno.writeTextFile(
  ASSISTANT_PATH,
  JSON.stringify(assistant, null, 2) + "\n",
);

console.log("Synced riley-assistant.json from core tool registry.");
