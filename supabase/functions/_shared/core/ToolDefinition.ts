import type { ParametersSchema } from "./types.ts";

export type ToolExecute = (parameters: unknown) => Promise<unknown>;

/**
 * One class for all tools: each operation is an instance with name, docs, schema, and execute.
 */
export class Tool {
  constructor(
    readonly name: string,
    readonly description: string,
    readonly manual: string,
    readonly parametersSchema: ParametersSchema,
    readonly execute: ToolExecute,
  ) {}
}
