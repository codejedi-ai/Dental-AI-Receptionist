// @ts-nocheck
// Follow this setup guide to integrate the Deno language server with your editor:
// https://deno.land/manual/getting_started/setup_your_environment
import defaultFunctions from "../_shared/function-call/dentalTools.ts";
import { normalizeOperationPayload } from "../_shared/core/lib/dispatchPayload.ts";
import { commonHeaders } from "../_shared/headers.ts";
import { VapiWebhookEnum } from "../_shared/vapi.types.ts";

const ROUTER_TOOL_NAME = "dispatch_dental_action";

const normalizeResult = (output) => {
  if (output === null || output === undefined) {
    return "";
  }
  if (typeof output === "string") {
    return output;
  }
  if (typeof output === "object") {
    const maybeObject = output;
    if (typeof maybeObject.result === "string") {
      return maybeObject.result;
    }
    return JSON.stringify(maybeObject);
  }
  return String(output);
};

/** Vapi requires `result` / `error` as plain strings; line breaks in `result` break parsing. */
function sanitizeVapiResultString(s) {
  return String(s).replace(/\r\n|\r|\n/g, " ").replace(/\s+/g, " ").trim();
}

/** Vapi sends OpenAI-style `arguments` as a JSON string; Chat may send `parameters` as an object. */
function parseArgs(raw) {
  if (raw == null || raw === undefined) return {};
  if (typeof raw === "string") {
    try {
      return JSON.parse(raw || "{}");
    } catch {
      return {};
    }
  }
  if (typeof raw === "object") return raw;
  return {};
}

function parseFunctionCallArguments(fc) {
  if (fc == null) return {};
  if (fc.parameters != null && typeof fc.parameters === "object") return fc.parameters;
  return parseArgs(fc.arguments);
}

/** Same routing as legacy function-call: single router tool + direct tools. */
async function executeToolInvocation(name, parameters, functions = defaultFunctions) {
  if (name === ROUTER_TOOL_NAME || name === "dental_tool") {
    const input = parameters ?? {};
    const routedName = (
      input.operation ?? input.toolName ?? input.action ?? ""
    ).trim();
    let routedParams = input.payload ?? input.input ??
      input.parameters ?? input.args ?? {};
    if (routedParams == null || typeof routedParams !== "object") {
      routedParams = {};
    }

    if (input.requestId) {
      console.log(`[dispatch] requestId=${String(input.requestId)} operation=${routedName}`);
    }

    if (!routedName) {
      return {
        result:
          `Missing operation for ${ROUTER_TOOL_NAME} (e.g., operation: "get_current_date").`,
      };
    }

    const routedFn = functions[routedName];
    if (!routedFn) {
      return { result: `Unknown routed tool: ${routedName}` };
    }
    const normalized = normalizeOperationPayload(routedName, routedParams);
    return await routedFn(normalized);
  }

  const fn = functions[name];
  if (!fn) {
    return { result: `Unknown function: ${name}` };
  }
  return await fn(parameters);
}

/**
 * Vapi tool execution expects HTTP 200 and body:
 * { "results": [ { "toolCallId": "<id from request>", "result": "<string>" } ] }
 * See https://docs.vapi.ai/server-url/events (tool-calls).
 */
function toolResultsResponse(results) {
  const safe = results.map((r) => ({
    toolCallId: r.toolCallId,
    ...(r.error != null
      ? { error: sanitizeVapiResultString(r.error) }
      : { result: sanitizeVapiResultString(r.result ?? "") }),
  }));
  return new Response(JSON.stringify({ results: safe }), {
    status: 200,
    headers: commonHeaders,
  });
}

/**
 * Chat / custom-tool POST bodies sometimes omit `message` and send toolCallList or functionCall at the root.
 * If we only read `reqBody.message`, payload is undefined and we incorrectly return `{}` → "No result returned".
 */
function extractWebhookPayload(reqBody) {
  if (!reqBody || typeof reqBody !== "object") return {};
  const inner = reqBody.message;
  const hasInner =
    inner != null &&
    typeof inner === "object" &&
    (Array.isArray(inner.toolCallList) && inner.toolCallList.length > 0 ||
      Array.isArray(inner.tool_calls) && inner.tool_calls.length > 0 ||
      Array.isArray(inner.toolWithToolCallList) && inner.toolWithToolCallList.length > 0 ||
      inner.functionCall != null ||
      inner.type != null);
  if (hasInner) return inner;
  if (
    reqBody.toolCallList?.length ||
    reqBody.tool_calls?.length ||
    reqBody.toolWithToolCallList?.length ||
    reqBody.functionCall
  ) {
    return reqBody;
  }
  if (inner != null && typeof inner === "object") return inner;
  return reqBody;
}

/**
 * OpenAI / Chat often sends tool calls as:
 * { "id": "call_…", "type": "function", "function": { "name": "…", "arguments": "{}" } }
 * with NO top-level `name`. We must read function.name and function.arguments.
 */
function flattenToolCallForExecution(raw) {
  if (!raw || typeof raw !== "object") {
    return { id: undefined, name: undefined, args: {} };
  }
  const fn = raw.function;
  const id =
    raw.id ??
    raw.toolCallId ??
    raw.tool_call_id ??
    (fn && typeof fn === "object" ? fn.toolCallId : undefined);
  const name =
    (typeof raw.name === "string" && raw.name
      ? raw.name
      : undefined) ??
    (fn && typeof fn.name === "string" ? fn.name : undefined);
  const argsRaw =
    raw.arguments ??
    raw.parameters ??
    (fn && (fn.arguments ?? fn.parameters));
  const args = parseArgs(argsRaw);
  return { id, name, args };
}

/** Normalize tool-calls payloads (toolCallList vs toolWithToolCallList vs tool_calls). */
function normalizeToolCallList(payload) {
  if (!payload) return null;
  if (Array.isArray(payload.toolCallList) && payload.toolCallList.length > 0) {
    return payload.toolCallList;
  }
  if (Array.isArray(payload.tool_calls) && payload.tool_calls.length > 0) {
    return payload.tool_calls;
  }
  if (Array.isArray(payload.toolWithToolCallList) && payload.toolWithToolCallList.length > 0) {
    return payload.toolWithToolCallList.map((t) => ({
      id: t.toolCall?.id,
      name: t.name,
      arguments: t.toolCall?.parameters ?? t.toolCall?.arguments,
    }));
  }
  return null;
}

async function runToolCallList(toolCallList) {
  const results = await Promise.all(toolCallList.map(async (raw) => {
    const { id, name, args } = flattenToolCallForExecution(raw);
    if (!id || !name) {
      return {
        toolCallId: id ?? "unknown",
        result:
          "Invalid tool call payload: need id (call_…) and function name; Chat sends function.name + function.arguments.",
      };
    }
    try {
      const output = await executeToolInvocation(name, args);
      return {
        toolCallId: id,
        result: normalizeResult(output),
      };
    } catch (error) {
      return {
        toolCallId: id,
        result: `Tool execution failed for ${name}: ${error instanceof Error ? error.message : "Unknown error"}`,
      };
    }
  }));
  return results;
}

async function endOfCallReportHandler(payload) {
  console.log("end-of-call-report", {
    endedReason: payload.endedReason,
    summary: payload.summary,
    recordingUrl: payload.recordingUrl,
  });
}

async function statusUpdateHandler(_payload) {
  return {};
}

async function speechUpdateHandler(_payload) {
  return {};
}

async function transcriptHandler(_payload) {
  return {};
}

async function hangEventHandler(_payload) {
  return {};
}

Deno.serve(async (req) => {
  if (req.method === "GET") {
    return new Response(JSON.stringify({
      ok: true,
      status: "ok",
      service: "webhook",
    }), {
      status: 200,
      headers: commonHeaders,
    });
  }
  if (req.method === "POST") {
    const reqBody = await req.json();
    try {
      const payload = extractWebhookPayload(reqBody);

      const normalizedList = normalizeToolCallList(payload);
      if (normalizedList) {
        const results = await runToolCallList(normalizedList);
        return toolResultsResponse(results);
      }

      const msgType = payload?.type;
      const fc = payload?.functionCall;
      const hasFunctionCall = fc != null && typeof fc === "object";
      if (
        msgType === "tool-calls" ||
        msgType === VapiWebhookEnum.FUNCTION_CALL ||
        msgType === "function-call" ||
        hasFunctionCall
      ) {
        if (!fc) {
          throw new Error("Invalid request: missing functionCall.");
        }
        const name = fc.name;
        const params = parseFunctionCallArguments(fc);
        const toolCallId =
          fc.id ?? fc.toolCallId ?? payload.toolCallId ?? "";
        const output = await executeToolInvocation(name, params);
        return toolResultsResponse([{
          toolCallId: toolCallId || "unknown",
          result: normalizeResult(output),
        }]);
      }

      switch (msgType) {
        case VapiWebhookEnum.STATUS_UPDATE: {
          return new Response(JSON.stringify(await statusUpdateHandler(payload)), {
            status: 201,
            headers: commonHeaders,
          });
        }
        case VapiWebhookEnum.ASSISTANT_REQUEST:
          return new Response(JSON.stringify({}), {
            status: 200,
            headers: commonHeaders,
          });
        case VapiWebhookEnum.END_OF_CALL_REPORT:
          return new Response(JSON.stringify(await endOfCallReportHandler(payload)), {
            status: 201,
            headers: commonHeaders,
          });
        case VapiWebhookEnum.SPEECH_UPDATE:
          return new Response(JSON.stringify(await speechUpdateHandler(payload)), {
            status: 201,
            headers: commonHeaders,
          });
        case VapiWebhookEnum.TRANSCRIPT:
          return new Response(JSON.stringify(await transcriptHandler(payload)), {
            status: 201,
            headers: commonHeaders,
          });
        case VapiWebhookEnum.HANG:
          return new Response(JSON.stringify(await hangEventHandler(payload)), {
            status: 201,
            headers: commonHeaders,
          });
        default:
          return new Response(JSON.stringify({}), {
            status: 200,
            headers: commonHeaders,
          });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      return new Response(JSON.stringify({
        error: message,
      }), {
        status: 500,
        headers: {
          ...commonHeaders,
        },
      });
    }
  }
  return new Response(JSON.stringify({
    message: "Not Found",
  }), {
    status: 404,
    headers: commonHeaders,
  });
}); /* To invoke locally:

  1. Run `supabase start` (see: https://supabase.com/docs/reference/cli/supabase-start)
  2. Make an HTTP request:

  curl -i --location --request POST 'http://127.0.0.1:54321/functions/v1/webhook'     --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0'     --header 'Content-Type: application/json'     --data '{"name":"Functions"}'

*/
