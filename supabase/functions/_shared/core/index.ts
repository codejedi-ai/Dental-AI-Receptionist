import { SupabaseTool } from "./SupabaseTool.ts";
import { Tool } from "./ToolDefinition.ts";
import { book_appointment } from "./dental_action/book_appointment.ts";
import { cancel_appointment } from "./dental_action/cancel_appointment.ts";
import { check_availability } from "./dental_action/check_availability.ts";
import { get_clinic_info } from "./dental_action/get_clinic_info.ts";
import { get_current_date } from "./dental_action/get_current_date.ts";
import { get_dentists } from "./dental_action/get_dentists.ts";
import { get_next_available_dates } from "./dental_action/get_next_available_dates.ts";
import { health_check } from "./dental_action/health_check.ts";
import { parse_date } from "./dental_action/parse_date.ts";
import { send_booking_confirmation } from "./dental_action/send_booking_confirmation.ts";

const allTools: Tool[] = [
  health_check,
  get_current_date,
  get_clinic_info,
  get_dentists,
  parse_date,
  check_availability,
  get_next_available_dates,
  book_appointment,
  cancel_appointment,
  send_booking_confirmation,
];

/** Single registry: handlers, router copy, and bundled manuals. */
export const supabaseTool = new SupabaseTool(allTools);

export const defaultToolHandlers = supabaseTool.getHandlers();

export { Tool, SupabaseTool };
export type { ParametersSchema } from "./types.ts";
