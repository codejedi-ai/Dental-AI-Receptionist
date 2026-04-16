/** OpenAI-compatible JSON Schema fragment for one property. */
export type ParameterProperty = {
  type: string;
  description: string;
  enum?: string[];
  items?: ParameterProperty;
  properties?: Record<string, ParameterProperty>;
};

/** Full parameters object schema for a tool. */
export type ParametersSchema = {
  type: "object";
  properties: Record<string, ParameterProperty>;
  required?: string[];
};
