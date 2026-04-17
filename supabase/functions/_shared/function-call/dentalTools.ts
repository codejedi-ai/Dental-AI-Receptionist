/**
 * Dental / scheduling tool implementations live under `_shared/core/` (one class per tool).
 * This module re-exports the flat handler map expected by the webhook.
 */
import {
  defaultToolHandlers,
  supabaseTool,
} from "../core/index.ts";

export default defaultToolHandlers;

export const {
  health_check,
  get_current_date,
  get_current_datetime,
  get_clinic_info,
  get_dentists,
  parse_date,
  get_next_available_dates,
  check_availability,
  book_appointment,
  cancel_appointment,
  send_booking_confirmation,
} = defaultToolHandlers;

export { supabaseTool };
