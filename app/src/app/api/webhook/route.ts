import { NextRequest, NextResponse } from "next/server";
import { getDb, getAvailableSlots } from "@/lib/db";
import { clinicConfig } from "@/lib/clinic-config";
import { formatTime } from "@/lib/utils";

export async function POST(req: NextRequest) {
  const body = await req.json();

  // Vapi sends tool call messages
  const { message } = body;

  if (!message || message.type !== "function-call") {
    return NextResponse.json({ error: "Not a function call" }, { status: 400 });
  }

  const { functionCall } = message;
  const { name, parameters } = functionCall;

  try {
    switch (name) {
      case "check_availability": {
        const { date, dentist } = parameters;
        const dentistName = dentist || clinicConfig.dentists[0].name;
        const slots = getAvailableSlots(date, dentistName);

        if (slots.length === 0) {
          return NextResponse.json({
            results: [
              {
                result: `Sorry, there are no available slots on ${date} with ${dentistName}. Would you like to try a different date or dentist?`,
              },
            ],
          });
        }

        const formattedSlots = slots.slice(0, 6).map(formatTime).join(", ");
        return NextResponse.json({
          results: [
            {
              result: `Available times on ${date} with ${dentistName}: ${formattedSlots}${
                slots.length > 6 ? ` and ${slots.length - 6} more slots` : ""
              }. Which time works best for you?`,
            },
          ],
        });
      }

      case "book_appointment": {
        const { date, time, dentist, service, patient_name, patient_phone, patient_email, notes } = parameters;
        const db = getDb();

        // Find or create patient
        let patient: any = null;
        if (patient_email) {
          patient = db.prepare("SELECT * FROM patients WHERE email = ?").get(patient_email);
        }
        if (!patient && patient_name) {
          patient = db.prepare("SELECT * FROM patients WHERE name = ? AND phone = ?").get(patient_name, patient_phone || "");
        }
        if (!patient) {
          const result = db
            .prepare("INSERT INTO patients (name, email, phone) VALUES (?, ?, ?)")
            .run(patient_name || "Voice Patient", patient_email || "", patient_phone || "");
          patient = { id: result.lastInsertRowid };
        }

        const dentistName = dentist || clinicConfig.dentists[0].name;
        const serviceName = service || "Checkup";

        // Verify slot is still available
        const availableSlots = getAvailableSlots(date, dentistName);
        if (!availableSlots.includes(time)) {
          return NextResponse.json({
            results: [
              {
                result: `Sorry, the ${formatTime(time)} slot on ${date} is no longer available. Would you like to pick another time?`,
              },
            ],
          });
        }

        const result = db
          .prepare(
            "INSERT INTO appointments (patient_id, dentist, service, date, time, status, notes) VALUES (?, ?, ?, ?, ?, 'confirmed', ?)"
          )
          .run(patient.id, dentistName, serviceName, date, time, notes || "Booked via voice assistant");

        return NextResponse.json({
          results: [
            {
              result: `I've booked your ${serviceName} appointment with ${dentistName} on ${date} at ${formatTime(time)}. Your appointment number is ${result.lastInsertRowid}. Is there anything else I can help you with?`,
            },
          ],
        });
      }

      case "cancel_appointment": {
        const { appointment_id, patient_name } = parameters;
        const db = getDb();

        let appointment: any = null;
        if (appointment_id) {
          appointment = db.prepare("SELECT * FROM appointments WHERE id = ? AND status = 'confirmed'").get(appointment_id);
        } else if (patient_name) {
          appointment = db
            .prepare(
              `SELECT a.* FROM appointments a
               JOIN patients p ON a.patient_id = p.id
               WHERE p.name LIKE ? AND a.status = 'confirmed'
               ORDER BY a.date ASC LIMIT 1`
            )
            .get(`%${patient_name}%`);
        }

        if (!appointment) {
          return NextResponse.json({
            results: [
              {
                result: "I couldn't find an active appointment with that information. Could you provide your appointment number or full name?",
              },
            ],
          });
        }

        db.prepare("UPDATE appointments SET status = 'cancelled' WHERE id = ?").run(appointment.id);

        return NextResponse.json({
          results: [
            {
              result: `Your appointment #${appointment.id} on ${appointment.date} at ${formatTime(appointment.time)} has been cancelled. Would you like to rebook?`,
            },
          ],
        });
      }

      case "get_clinic_info": {
        return NextResponse.json({
          results: [
            {
              result: `${clinicConfig.name} is located at ${clinicConfig.address}. Our phone number is ${clinicConfig.phone}. We're open Monday to Friday ${clinicConfig.hours.weekdays}, Saturday ${clinicConfig.hours.saturday}, and closed on Sunday. Our dentists are ${clinicConfig.dentists.map((d) => `${d.name} (${d.specialty})`).join(", ")}.`,
            },
          ],
        });
      }

      case "get_services": {
        const serviceList = clinicConfig.services
          .map((s) => `${s.name} (${s.price})`)
          .join(", ");
        return NextResponse.json({
          results: [
            {
              result: `We offer the following services: ${serviceList}. Which service are you interested in?`,
            },
          ],
        });
      }

      default:
        return NextResponse.json({
          results: [{ result: "I'm not sure how to help with that. Would you like to book an appointment or check availability?" }],
        });
    }
  } catch (error) {
    console.error("Webhook error:", error);
    return NextResponse.json({
      results: [{ result: "I'm sorry, something went wrong. Please try again or call us directly." }],
    });
  }
}
