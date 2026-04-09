package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	_ "time/tzdata" // embed timezone DB

	"dental-ai-vapi/internal/db"
	"dental-ai-vapi/internal/util"
)

var weekdaySlots = []string{"08:00", "08:30", "09:00", "09:30", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "13:30", "14:00", "14:30", "15:00", "15:30", "16:00", "16:30", "17:00", "17:30"}
var saturdaySlots = []string{"09:00", "09:30", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "13:30"}

func CheckAvailability(ctx context.Context, pg *db.Postgres, args json.RawMessage) (string, string) {
	date := util.ParseArg(args, "date")
	if date == "" {
		date = util.ParseArg(args, "appointmentDate")
	}
	dentist := util.ParseArg(args, "dentist")
	if dentist == "any" {
		dentist = ""
	}

	if strings.TrimSpace(date) == "" {
		return "I can check availability once I have a date in YYYY-MM-DD format, for example 2026-04-15.", "error"
	}

	log.Printf("[CheckAvailability] date=%q dentist=%q raw_args=%s", date, dentist, string(args))

	requestedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Printf("[CheckAvailability] Invalid date format: %s", date)
		return fmt.Sprintf("Invalid date format: %s. Please use YYYY-MM-DD.", date), "error"
	}

	today := time.Now().Truncate(24 * time.Hour)
	if requestedDate.Before(today) {
		log.Printf("[CheckAvailability] Date %s is in the past", date)
		return "That date is in the past. Could you provide a future date?", "error"
	}

	dow := requestedDate.Weekday()
	if dow == time.Sunday {
		log.Printf("[CheckAvailability] Date %s is a Sunday", date)
		return "Sorry, the clinic is closed on Sundays. We're open Monday–Friday 8 AM–6 PM and Saturday 9 AM–2 PM.", "error"
	}

	slots := weekdaySlots
	if dow == time.Saturday {
		slots = saturdaySlots
	}

	// Load dentists from DB so the list stays in sync with the database
	allDentists, err := pg.GetAllDentists(ctx)
	if err != nil || len(allDentists) == 0 {
		log.Printf("[CheckAvailability] GetAllDentists error: %v", err)
		return "Sorry, I'm having trouble accessing the dentist list right now. Please try again.", "error"
	}
	log.Printf("[CheckAvailability] Loaded %d dentists from DB: %v", len(allDentists), allDentists)

	dentistList := allDentists
	if dentist != "" {
		dentistList = []string{dentist}
		log.Printf("[CheckAvailability] Filtering to specific dentist: %s", dentist)
	}

	var available []struct{ Date, Dentist, Time string }

	for _, d := range dentistList {
		did, err := pg.GetDentistID(ctx, d)
		if err != nil {
			log.Printf("[CheckAvailability] GetDentistID error for %q: %v", d, err)
			continue
		}
		log.Printf("[CheckAvailability] Dentist %q has ID %d", d, did)

		booked, err := pg.GetBookedSlots(ctx, date, &did)
		if err != nil {
			log.Printf("[CheckAvailability] GetBookedSlots error for date=%s dentist_id=%d: %v", date, did, err)
			continue
		}
		log.Printf("[CheckAvailability] Found %d booked slots for %s dentist_id=%d", len(booked), date, did)

		bookedTimes := make(map[string]bool)
		for _, b := range booked {
			bookedTimes[b.Time] = true
		}

		for _, slot := range slots {
			if !bookedTimes[slot] {
				available = append(available, struct{ Date, Dentist, Time string }{date, d, slot})
			}
		}
	}

	if len(available) == 0 {
		msg := fmt.Sprintf("No available slots on %s.", date)
		if dentist != "" {
			msg = fmt.Sprintf("No available slots on %s with %s.", date, dentist)
		}
		msg += " Would you like to try another date?"
		return msg, "error"
	}

	limit := 5
	if len(available) < limit {
		limit = len(available)
	}

	var parts []string
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("%s with %s", util.TimeString(util.MustParseTime(available[i].Time)), available[i].Dentist))
	}

	result := fmt.Sprintf("Available slots on %s: %s.", date, util.JoinStr(parts, ", "))
	if len(available) > 5 {
		result += fmt.Sprintf(" (%d more available)", len(available)-5)
	}
	result += " Which time works for you?"

	return result, "success"
}

func BookAppointment(ctx context.Context, pg *db.Postgres, mongo *db.Mongo, args json.RawMessage) (string, string) {
	patientName := util.ParseArg(args, "patientName")
	patientPhone := util.ParseArg(args, "patientPhone")
	patientEmail := util.ParseArgOpt(args, "patientEmail")
	date := util.ParseArg(args, "date")
	timeStr := util.ParseArg(args, "time")
	dentist := util.ParseArg(args, "dentist")
	service := util.ParseArg(args, "service")
	notes := util.ParseArgOpt(args, "notes")

	// STRICT IDENTITY: Only full name + phone required. Email is fallback if no phone.
	if patientName == "" || date == "" || timeStr == "" || dentist == "" {
		return "I'm missing details. I need the patient's full name, date, time, and dentist to book.", "error"
	}
	if patientPhone == "" && patientEmail == nil {
		return "I need either your mobile number or email address for booking confirmation. Which do you prefer?", "error"
	}

	if service == "" {
		service = "General Checkup"
	}
	service = normalizeServiceName(service)

	patientID, err := pg.FindOrCreatePatient(ctx, patientName, patientPhone, patientEmail, nil)
	if err != nil {
		log.Printf("FindOrCreatePatient error: %v", err)
		return "Sorry, I had trouble saving your information. Please try again.", "error"
	}

	dentistID, err := pg.GetDentistID(ctx, dentist)
	if err != nil {
		log.Printf("GetDentistID error for %q: %v", dentist, err)
		return fmt.Sprintf("Dentist '%s' not found.", dentist), "error"
	}

	appt, err := pg.CreateAppointment(ctx, patientID, dentistID, service, date, timeStr, notes)
	if err != nil {
		log.Printf("CreateAppointment error: %v", err)
		return "Sorry, I had trouble booking that appointment. The slot may be taken. Please try a different time.", "error"
	}

	extra := ""
	if patientEmail != nil {
		extra = fmt.Sprintf(" We'll follow up at %s if needed.", *patientEmail)
	}

	msg := fmt.Sprintf("Booked! %s has a %s appointment with %s on %s at %s. Appointment #%s.%s Please arrive 15 minutes early and bring your insurance card. Anything else?",
		appt.PatientName, appt.Service, appt.Dentist, appt.Date,
		util.TimeString(util.MustParseTime(appt.Time)), appt.ID, extra)

	return msg, "success"
}

func normalizeServiceName(service string) string {
	s := strings.ToLower(strings.TrimSpace(service))
	serviceMap := map[string]string{
		"bridge":       "Bridge",
		"cleaning":     "Cleaning",
		"consultation": "Consultation",
		"consultaion":  "Consultation",
		"crown":        "Crown",
	}
	if normalized, ok := serviceMap[s]; ok {
		return normalized
	}
	if service == "" {
		return service
	}
	return strings.TrimSpace(service)
}

func CancelAppointment(ctx context.Context, pg *db.Postgres, mongo *db.Mongo, args json.RawMessage) (string, string) {
	patientName := util.ParseArg(args, "patientName")
	date := util.ParseArg(args, "date")

	if patientName == "" || date == "" {
		return "I need the patient name and date to find the appointment. Could you provide those?", "error"
	}

	appt, err := pg.FindAppointment(ctx, patientName, date)
	if err != nil {
		return fmt.Sprintf("I couldn't find an appointment for %s on %s. Could you double-check the name and date?", patientName, date), "error"
	}

	apptID := 0
	fmt.Sscanf(appt.ID, "%d", &apptID)

	cancelled, err := pg.CancelAppointment(ctx, apptID)
	if err != nil || cancelled == nil {
		return "That appointment was already cancelled or not found.", "error"
	}

	msg := fmt.Sprintf("Cancelled! %s's %s on %s at %s has been removed. Would you like to rebook?",
		cancelled.PatientName, cancelled.Service, cancelled.Date,
		util.TimeString(util.MustParseTime(cancelled.Time)))

	return msg, "success"
}

func GetClinicInfo(ctx context.Context, mongo *db.Mongo, args json.RawMessage) (string, string) {
	topic := util.ParseArg(args, "topic")
	if topic == "" {
		topic = "general"
	}
	info := mongo.GetClinicInfo(ctx, topic)
	return info, "success"
}




func LookupPatient(ctx context.Context, pg *db.Postgres, args json.RawMessage) (string, string) {
	phone := util.ParseArg(args, "phone")
	if phone == "" {
		return "I need the phone number to look up the patient.", "error"
	}
	pt, err := pg.GetPatientByPhone(ctx, phone)
	if err != nil {
		return "No existing record found. This is a new patient — please collect their full name and mobile number.", "success"
	}
	// STRICT: return ONLY name + phone. Never leak address, email, or DOB to the AI.
	info := fmt.Sprintf("Returning patient found — Name: %s, Phone: %s, UUID: %s", pt.Name, pt.Phone, pt.UUID)
	return info, "success"
}

func SendBookingConfirmation(args json.RawMessage) (string, string) {
	phone := util.ParseArg(args, "phone")
	name := util.ParseArg(args, "patientName")
	service := util.ParseArg(args, "service")
	dentist := util.ParseArg(args, "dentist")
	date := util.ParseArg(args, "date")
	apptTime := util.ParseArg(args, "time")

	if phone == "" || name == "" {
		return "Missing phone or patient name for confirmation.", "error"
	}

	// Format date nicely if possible
	if t, err := time.Parse("2006-01-02", date); err == nil {
		date = t.Format("Monday, January 2, 2006")
	}
	// Format time
	if t, err := time.Parse("15:04", apptTime); err == nil {
		apptTime = t.Format("3:04 PM")
	}

	msg := fmt.Sprintf(
		"Appointment booked for %s — %s with %s on %s at %s at Smile Dental Clinic. "+
			"123 Main Street, Newmarket, ON. Please arrive 15 min early.",
		name, service, dentist, date, apptTime,
	)
	log.Printf("Booking confirmation details for %s (phone on file: %s)", name, phone)
	return msg, "success"
}

func GetDentists(ctx context.Context, pg *db.Postgres) (string, string) {
	names, err := pg.GetAllDentists(ctx)
	if err != nil || len(names) == 0 {
		return "I'm having trouble retrieving the dentist list right now.", "error"
	}
	return fmt.Sprintf("Our dentists are: %s.", util.JoinStr(names, ", ")), "success"
}

func GetCurrentDate() (string, string) {
	now := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	now = now.In(loc)
	date := now.Format("Monday, January 2, 2006")
	timeStr := now.Format("3:04 PM")
	return fmt.Sprintf("Today is %s, %s Toronto time.", date, timeStr), "success"
}

// ValidatePatientInfo enforces the strict rule: only name + cell phone are required.
// Email is collected ONLY if the patient has no mobile number.
// Returns validation result and any missing fields.
func ValidatePatientInfo(args json.RawMessage) (string, string) {
	name := util.ParseArg(args, "patient_name")
	phone := util.ParseArg(args, "patient_phone")
	email := util.ParseArgOpt(args, "patient_email")

	// Name is ALWAYS required
	if name == "" {
		return `{"valid": false, "missing": ["patient_name"], "message": "I need your full name to book the appointment."}`, "error"
	}

	// Phone is required. If caller explicitly says they don't have a phone, email becomes required.
	noPhone := util.ParseArg(args, "no_mobile") == "true"

	if phone == "" && !noPhone {
		return `{"valid": false, "missing": ["patient_phone"], "message": "I need your cell phone number for our records."}`, "error"
	}

	if phone == "" && noPhone {
		// Patient has no mobile — email is now required
		if email == nil || *email == "" {
			return `{"valid": false, "missing": ["patient_email"], "message": "Since you don't have a mobile number, I'll need your email address to send the booking confirmation."}`, "error"
		}
		return `{"valid": true, "method": "email_confirmation", "message": "Thank you. I'll send the confirmation to your email."}`, "success"
	}

	// Normal case: name + phone collected
	return `{"valid": true, "method": "phone_confirmation", "message": "Thank you. I have your name and phone on file."}`, "success"
}

// ParseDate converts a natural language date to YYYY-MM-DD format.
// Supports: "today", "tomorrow", "next Monday", "this Friday", "March 15", etc.
func ParseDate(ctx context.Context, args json.RawMessage) (string, string) {
	input := util.ParseArg(args, "date_text")
	if input == "" {
		return `{"error": "date_text is required"}`, "error"
	}

	now := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	now = now.In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	lower := strings.ToLower(strings.TrimSpace(input))

	// Relative dates
	switch {
	case lower == "today" || lower == "今天":
		return fmt.Sprintf(`{"date": "%s", "display": "today"}`, today.Format("2006-01-02")), "success"
	case lower == "tomorrow" || lower == "明天":
		t := today.AddDate(0, 0, 1)
		return fmt.Sprintf(`{"date": "%s", "display": "tomorrow (%s)"}`, t.Format("2006-01-02"), t.Format("Monday")), "success"
	case strings.HasPrefix(lower, "next "):
		dayName := strings.TrimPrefix(lower, "next ")
		if dow, ok := parseDayOfWeek(dayName); ok {
			daysAhead := (int(dow) - int(today.Weekday()) + 7) % 7
			if daysAhead == 0 {
				daysAhead = 7 // "next Monday" when today is Monday → next week
			}
			t := today.AddDate(0, 0, daysAhead)
			return fmt.Sprintf(`{"date": "%s", "display": "%s"}`, t.Format("2006-01-02"), t.Format("Monday, January 2")), "success"
		}
	case strings.HasPrefix(lower, "this "):
		dayName := strings.TrimPrefix(lower, "this ")
		if dow, ok := parseDayOfWeek(dayName); ok {
			daysAhead := (int(dow) - int(today.Weekday()) + 7) % 7
			if daysAhead == 0 {
				daysAhead = 0 // "this Monday" when today is Monday → today
			}
			t := today.AddDate(0, 0, daysAhead)
			return fmt.Sprintf(`{"date": "%s", "display": "%s"}`, t.Format("2006-01-02"), t.Format("Monday, January 2")), "success"
		}
	}

	// Absolute date: "March 15", "January 5th", "2025-03-15"
	// Try YYYY-MM-DD first
	if t, err := time.Parse("2006-01-02", lower); err == nil {
		return fmt.Sprintf(`{"date": "%s", "display": "%s"}`, t.Format("2006-01-02"), t.Format("Monday, January 2")), "success"
	}

	// Try "January 15", "March 5th", etc.
	for _, layout := range []string{"January 2", "January 2nd", "January 2st", "January 2nd, 2006", "Jan 2", "Jan 2, 2006"} {
		if t, err := time.Parse(layout, lower); err == nil {
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
			}
			if t.Before(today) {
				t = t.AddDate(1, 0, 0) // Assume next year if in the past
			}
			return fmt.Sprintf(`{"date": "%s", "display": "%s"}`, t.Format("2006-01-02"), t.Format("Monday, January 2")), "success"
		}
	}

	// Chinese: "三月十五", "下周一"
	if strings.Contains(lower, "周") || strings.Contains(lower, "星期") {
		dayMap := map[string]time.Weekday{
			"周一": time.Monday, "星期一": time.Monday,
			"周二": time.Tuesday, "星期二": time.Tuesday,
			"周三": time.Wednesday, "星期三": time.Wednesday,
			"周四": time.Thursday, "星期四": time.Thursday,
			"周五": time.Friday, "星期五": time.Friday,
			"周六": time.Saturday, "星期六": time.Saturday,
			"周日": time.Sunday, "星期日": time.Sunday,
		}
		for zh, dow := range dayMap {
			if strings.Contains(lower, zh) {
				daysAhead := (int(dow) - int(today.Weekday()) + 7) % 7
				if daysAhead == 0 {
					daysAhead = 7
				}
				t := today.AddDate(0, 0, daysAhead)
				return fmt.Sprintf(`{"date": "%s", "display": "%s"}`, t.Format("2006-01-02"), t.Format("Monday, January 2")), "success"
			}
		}
	}

	return fmt.Sprintf(`{"error": "Could not parse date from: %s. Please say something like 'tomorrow', 'next Monday', or 'March 15'.", "input": %q}`, input, input), "error"
}

// GetNextAvailableDates returns the next N dates that have available slots.
// This helps the AI suggest dates when the caller asks "what's the earliest available?"
func GetNextAvailableDates(ctx context.Context, pg *db.Postgres, args json.RawMessage) (string, string) {
	daysToSearch := 14 // Look ahead 2 weeks by default
	if n := util.ParseArg(args, "days"); n != "" {
		if parsed, err := fmt.Sscanf(n, "%d", &daysToSearch); parsed == 0 || err != nil {
			daysToSearch = 14
		}
	}

	now := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	now = now.In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	allDentists, err := pg.GetAllDentists(ctx)
	if err != nil || len(allDentists) == 0 {
		return "Sorry, I'm having trouble accessing the schedule right now.", "error"
	}

	var availableDates []string

	for i := 1; i <= daysToSearch; i++ { // Start from tomorrow
		date := today.AddDate(0, 0, i)
		dow := date.Weekday()

		if dow == time.Sunday {
			continue // Closed Sundays
		}

		dateStr := date.Format("2006-01-02")
		dayName := date.Format("Monday")

		// Count total available slots across all dentists
		totalSlots := 0
		for _, dentist := range allDentists {
			did, err := pg.GetDentistID(ctx, dentist)
			if err != nil {
				continue
			}

			slots := weekdaySlots
			if dow == time.Saturday {
				slots = saturdaySlots
			}

			booked, err := pg.GetBookedSlots(ctx, dateStr, &did)
			if err != nil {
				continue
			}

			bookedTimes := make(map[string]bool)
			for _, b := range booked {
				bookedTimes[b.Time] = true
			}

			for _, slot := range slots {
				if !bookedTimes[slot] {
					totalSlots++
				}
			}
		}

		if totalSlots > 0 {
			availableDates = append(availableDates, fmt.Sprintf("%s (%d slots)", dayName, totalSlots))
			if len(availableDates) >= 5 {
				break // Return top 5 dates
			}
		}
	}

	if len(availableDates) == 0 {
		return "I don't see any available slots in the next two weeks. Would you like me to put you on a waitlist or check a specific date?", "error"
	}

	return fmt.Sprintf("The next available dates are: %s. Which would you prefer?", util.JoinStr(availableDates, ", ")), "success"
}

// parseDayOfWeek converts a day name to time.Weekday.
func parseDayOfWeek(s string) (time.Weekday, bool) {
	days := map[string]time.Weekday{
		"monday": time.Monday, "tuesday": time.Tuesday, "wednesday": time.Wednesday,
		"thursday": time.Thursday, "friday": time.Friday, "saturday": time.Saturday, "sunday": time.Sunday,
		"mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday,
		"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday, "sun": time.Sunday,
	}
	if dow, ok := days[s]; ok {
		return dow, true
	}
	return 0, false
}
